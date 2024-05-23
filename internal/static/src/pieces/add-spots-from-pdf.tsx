import { useCallback, useRef, useState } from "preact/hooks";
import { type PageImage, type CroppedImageData } from "../common";
import { Link } from "../ui/links";
import { CombineMode, SaveMode, SelectMode } from "./pdf-spot-components";
import { type PDFDocumentProxy } from "pdfjs-dist";

const pdfjs = await import("pdfjs-dist");

pdfjs.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.mjs",
  import.meta.url,
).toString();

async function extractPageToImage(pdf: PDFDocumentProxy, i: number) {
  const page = await pdf.getPage(i + 1);
  const canvas = document.createElement("canvas");
  // standardizing around approximately 10 x 12 inches at 300 dpi
  canvas.width = 10 * 300;
  const ctx = canvas.getContext("2d");
  const viewport = page.getViewport({
    scale: canvas.width / page.getViewport({ scale: 1 }).width,
  });
  canvas.height = viewport.height;
  const renderContext = {
    canvasContext: ctx!,
    viewport,
  };
  await page.render(renderContext).promise;
  const data = canvas.toDataURL("image/png");
  return {
    src: data,
    alt: `page ${i + 1}`,
    id: data.substring(20, 50),
  };
}

type Mode = "add" | "select" | "combine" | "save";

export function AddSpotsFromPDF(props: { pieceid: string; csrf: string }) {
  const [mode, setMode] = useState<Mode>("add");
  const [pageImages, setPageImages] = useState<PageImage[]>([]);
  const [spotImagesByPage, setSpotImagesByPage] = useState<
    CroppedImageData[][]
  >([]);
  const [progressTotal, setProgressTotal] = useState(0);
  const [progressCurrent, setProgressCurrent] = useState(0);

  const fileFormRef = useRef<HTMLFormElement>(null!);

  const addPDFFile = useCallback(() => {
    const fd = new FormData(fileFormRef.current);
    const pdf = fd.get("pdf") as File;
    const reader = new FileReader();
    reader.readAsArrayBuffer(pdf);
    reader.onloadend = () => {
      pdfjs
        .getDocument(reader.result as ArrayBuffer)
        .promise.then((pdf) => {
          const pagePromises: Promise<PageImage>[] = [];
          setProgressTotal(pdf.numPages);
          setProgressCurrent(0);
          for (let i = 0; i < pdf.numPages; i++) {
            const pagePromise = extractPageToImage(pdf, i).then((info) => {
              setProgressCurrent((curr) => curr + 1);
              return info;
            });
            pagePromises.push(pagePromise);
          }
          return Promise.all(pagePromises);
        })
        .then((results) => {
          setPageImages(results);
          setSpotImagesByPage(Array(results.length).fill([]));
          setMode("select");
          setProgressTotal(0);
          setProgressCurrent(0);
        })
        .catch(console.error);
    };
  }, [setSpotImagesByPage]);

  const savePageSpots = useCallback(
    (page: number, newImages: CroppedImageData[]) => {
      setSpotImagesByPage((currentImages) => {
        const nextImages = [...currentImages];
        nextImages[page] = [...newImages];
        return nextImages;
      });
    },
    [],
  );

  const saveCombinedSpots = useCallback(
    (
      newSpot: CroppedImageData,
      page: number,
      index: number,
      replace: boolean,
      removePage: number,
      removeIdx: number,
    ) => {
      if (replace) {
        setSpotImagesByPage((currentImages) => {
          const nextImages = [...currentImages];
          nextImages[page][index] = newSpot;
          nextImages[removePage].splice(removeIdx, 1);
          return nextImages;
        });
      } else {
        setSpotImagesByPage((currentImages) => {
          const nextImages = [...currentImages];
          nextImages[page].push(newSpot);
          return nextImages;
        });
      }
    },
    [setSpotImagesByPage],
  );

  return (
    <>
      {mode === "add" && (
        <div className="flex w-full flex-col gap-4 rounded-lg bg-white p-4 shadow-sm shadow-neutral-900/30 sm:mx-auto sm:max-w-3xl">
          <header className="flex flex-col gap-1">
            <h3 className="text-3xl font-bold">Add Spots</h3>
            <p className="text-lg italic">
              Upload a PDF to select your spots from your music, or enter them
              manually.
            </p>
          </header>
          <section className="flex flex-col flex-wrap justify-start gap-4 xs:flex-row">
            <div>
              <form
                ref={fileFormRef}
                onSubmit={(e) => e.preventDefault()}
                className="w-full"
              >
                <input
                  type="file"
                  name="pdf"
                  accept="application/pdf"
                  max="100MB"
                  className="green focusable flex-shrink"
                  onChange={() => addPDFFile()}
                />
              </form>
              {!!progressTotal && (
                <>
                  <div>Processing your file</div>
                  <progress
                    max={progressTotal}
                    value={progressCurrent}
                    className="progress-rounded progress-violet-600 progress-bg-violet-200 m-0 w-full"
                  />
                </>
              )}
            </div>
            <Link
              className="action-button amber focusable flex-shrink-0"
              href={`/library/pieces/${props.pieceid}/spots/add-single`}
            >
              <span className="icon-[ph--keyboard-thin] -ml-1 size-6" />
              Enter Manually
            </Link>
          </section>
        </div>
      )}
      {mode === "select" && (
        <SelectMode
          goBack={() => setMode("add")}
          pageImages={pageImages}
          spotImagesByPage={spotImagesByPage}
          savePageSpots={savePageSpots}
          done={() => setMode("combine")}
        />
      )}
      {mode === "combine" && (
        <CombineMode
          goBack={() => {
            setMode("select");
            setSpotImagesByPage(Array(spotImagesByPage.length).fill([]));
          }}
          goOn={() => setMode("save")}
          spotImagesByPage={spotImagesByPage}
          saveCombinedSpots={saveCombinedSpots}
        />
      )}
      {mode === "save" && (
        <SaveMode
          goBack={() => {
            setMode("combine");
          }}
          pieceid={props.pieceid}
          csrf={props.csrf}
          spotImagesByPage={spotImagesByPage}
        />
      )}
    </>
  );
}
