import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { cn, type PageImage, type CroppedImageData } from "../common";
import Cropper, { type CropperSelection } from "cropperjs";
import * as htmx from "htmx.org/dist/htmx";

export function SelectMode(props: {
  pageImages: PageImage[];
  spotImagesByPage: CroppedImageData[][];
  savePageSpots: (page: number, spotImagesByPage: CroppedImageData[]) => void;
  done: () => void;
  goBack: () => void;
}) {
  const [currentPage, setCurrentPage] = useState(0);

  const nextPage = useCallback(() => {
    setCurrentPage((current) =>
      current < props.pageImages.length - 1 ? current + 1 : current,
    );
  }, [props.pageImages]);

  const prevPage = useCallback(() => {
    setCurrentPage((current) => (current > 0 ? current - 1 : current));
  }, []);

  return (
    <>
      <div className="flex w-full flex-col items-center">
        <button
          className="action-button amber focusable"
          onClick={props.goBack}
          type="button"
        >
          Change PDF File
        </button>
        {props.pageImages.map((image, i) => (
          <div
            key={image.id}
            className={`w-full flex-col items-center justify-center py-4  ${
              i === currentPage ? "flex" : "hidden"
            }`}
          >
            <SingleCropper
              src={image.src}
              alt={image.alt}
              saveImages={(images) => props.savePageSpots(i, images)}
              show={i === currentPage}
              totalPages={props.pageImages.length - 1}
              currentPage={i}
              nextPage={nextPage}
              prevPage={prevPage}
              done={props.done}
            />
          </div>
        ))}
      </div>
    </>
  );
}

async function makeImageData(
  el: CropperSelection,
  i: number,
  transformationMatrix: number[],
): Promise<CroppedImageData> {
  const canv = await el.$toCanvas({
    width: el.width * 3,
    height: el.height * 3,
  });
  const data = canv.toDataURL();
  const imgID = `${i}-${el.x}-${el.y}-${el.width}-${el.height}`;
  return {
    data,
    id: imgID,
    width: canv.width,
    height: canv.height,
    x: el.x,
    y: el.y,
    transformationMatrix,
  };
}

export function SingleCropper(props: {
  src: string;
  alt: string;
  saveImages: (newImages: CroppedImageData[]) => void;
  show: boolean;
  done: () => void;
  currentPage: number;
  prevPage: () => void;
  nextPage: () => void;
  totalPages: number;
}) {
  const [hasShown, setHasShown] = useState(props.show);
  const ref = useRef<HTMLImageElement>(null!);
  const cropperRef = useRef<Cropper>(null!);

  const center = useCallback(() => {
    cropperRef.current.getCropperImage()?.$center();
  }, []);

  useEffect(() => {
    if (props.show && !hasShown && !!cropperRef.current) {
      setHasShown(true);
      cropperRef.current.getCropperImage()?.$zoom(-3);
      center();
    }
  }, [props.show, setHasShown, center, hasShown]);

  useEffect(() => {
    if (cropperRef.current) {
      cropperRef.current.container.querySelector("cropper-canvas")?.remove();
    }
    cropperRef.current = new Cropper(ref.current, {
      template: `
      <cropper-canvas id="cropper" style="height: max(800px, 80dvh); width: 90dvw; margin-inline: auto;">
        <cropper-image
          rotatable="false"
        ></cropper-image>
        <cropper-shade hidden></cropper-shade>
        <cropper-handle action="select" plain></cropper-handle>
        <cropper-selection
          initial-coverage="0"
          movable
          resizable
          zoomable
          multiple
          keyboard
        >
          <cropper-crosshair centered></cropper-crosshair>
          <cropper-handle
            action="move"
            theme-color="rgba(255, 255, 255, 0.35)"
          ></cropper-handle>
          <cropper-handle action="n-resize"></cropper-handle>
          <cropper-handle action="e-resize"></cropper-handle>
          <cropper-handle action="s-resize"></cropper-handle>
          <cropper-handle action="w-resize"></cropper-handle>
          <cropper-handle action="ne-resize"></cropper-handle>
          <cropper-handle action="nw-resize"></cropper-handle>
          <cropper-handle action="se-resize"></cropper-handle>
          <cropper-handle action="sw-resize"></cropper-handle>
        </cropper-selection>
      </cropper-canvas>
    `,
    });
  }, [props.src]);

  const getSpotImages = useCallback(() => {
    const selections = cropperRef.current.getCropperSelections();
    const transformationMatrix = cropperRef.current
      .getCropperImage()
      ?.$getTransform();
    if (!transformationMatrix) {
      return;
    }
    if (!selections) {
      return;
    }
    const workers: Promise<CroppedImageData>[] = [];
    for (let i = 0; i < selections.length; i++) {
      const el = selections[i];
      if (el.width < 5 || el.height < 5) {
        continue;
      }
      workers.push(makeImageData(el, i, transformationMatrix));
    }
    Promise.all(workers)
      .then((imageData) => {
        props.saveImages(imageData);
      })
      .catch(console.error);
  }, [props]);

  const handlePrevPage = useCallback(() => {
    getSpotImages();
    props.prevPage();
  }, [getSpotImages, props]);

  const handleNextPage = useCallback(() => {
    getSpotImages();
    props.nextPage();
  }, [getSpotImages, props]);

  const handleDone = useCallback(() => {
    getSpotImages();
    props.done();
  }, [getSpotImages, props]);

  return (
    <div className="flex flex-col items-center gap-2 rounded-xl border border-neutral-400 bg-white p-2 shadow-sm shadow-neutral-900/30">
      <div className="flex w-full items-center justify-center gap-2 px-4">
        <div className="flex w-full flex-wrap justify-between gap-4">
          <div className="flex flex-wrap items-center justify-start gap-2 sm:max-w-xl sm:gap-8">
            <button
              disabled={props.currentPage === 0}
              onClick={handlePrevPage}
              className="action-button neutral focusable"
            >
              <span
                class="icon-[iconamoon--arrow-left-6-circle-thin] -ml-1 size-6"
                aria-hidden="true"
              />
              Previous
            </button>
            <h2 className="text-2xl font-bold">Page {props.currentPage + 1}</h2>
            <button
              disabled={props.currentPage === props.totalPages}
              onClick={handleNextPage}
              className="action-button neutral focusable"
            >
              Next
              <span
                class="icon-[iconamoon--arrow-right-6-circle-thin] -mr-1 size-6"
                aria-hidden="true"
              />
            </button>
          </div>
          <button
            onClick={handleDone}
            className="action-button green focusable"
          >
            Done
            <span
              class="icon-[iconamoon--check-circle-1-thin] -mr-1 size-6"
              aria-hidden="true"
            />
          </button>
        </div>
      </div>
      <div style="min-height:800px" className="border border-neutral-400">
        <img ref={ref} width="100%" src={props.src} alt={props.alt} />
      </div>
    </div>
  );
}

export function CombineMode(props: {
  spotImagesByPage: CroppedImageData[][];
  saveCombinedSpots: (
    newSpot: CroppedImageData,
    page: number,
    index: number,
    replace: boolean,
    removePage: number,
    removeIdx: number,
  ) => void;
  goBack: () => void;
  goOn: () => void;
}) {
  // image data, page, index
  const [spot1, setSpot1] = useState<[CroppedImageData, number, number] | null>(
    null,
  );
  const [spot2, setSpot2] = useState<[CroppedImageData, number, number] | null>(
    null,
  );

  const toggleCombine = useCallback(
    (spot: CroppedImageData, page: number, index: number) => {
      if (spot === spot1?.[0]) {
        setSpot1(null);
      } else if (spot === spot2?.[0]) {
        setSpot2(null);
      } else if (spot1 === null) {
        setSpot1([spot, page, index]);
      } else if (spot2 === null) {
        setSpot2([spot, page, index]);
      }
    },
    [spot1, spot2, setSpot1, setSpot2],
  );

  const saveCombineImage = useCallback(
    (data: string, height: number, width: number) => {
      if (spot1 && spot2) {
        props.saveCombinedSpots(
          { data, id: `combined-${spot1[0].id}-${spot2[0].id}`, height, width },
          spot1[1],
          spot1[2],
          true,
          spot2[1],
          spot2[2],
        );
        setSpot1(null);
        setSpot2(null);
      }
    },
    [props, spot1, spot2],
  );

  const cancelCombine = useCallback(() => {
    setSpot1(null);
    setSpot2(null);
  }, []);

  return (
    <div className="flex w-full flex-col">
      <div className="flex w-full flex-col items-center justify-center gap-2 py-4">
        <div className="flex w-full justify-between">
          <button
            onClick={props.goBack}
            className="action-button amber focusable"
          >
            <span
              class="icon-[iconamoon--arrow-left-6-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Back to Choose
          </button>
          <h3 className="text-3xl font-bold">Combine</h3>
          <button
            onClick={props.goOn}
            className="action-button green focusable"
          >
            <span
              class="icon-[iconamoon--check-circle-1-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Done
          </button>
        </div>
        {!spot1 && (
          <h4 className="text-xl font-medium">Select first spot to combine</h4>
        )}
        {!spot2 && !!spot1 && (
          <h4 className="text-xl font-medium">Select second spot to combine</h4>
        )}
        {!!spot1 && !!spot2 && (
          <h4 className="text-xl font-medium">Adjust alignment and save</h4>
        )}
      </div>
      {!!spot1 && !!spot2 && (
        <div className="flex w-full px-4">
          <CombineCanvas
            combineImageData={[spot1[0], spot2[0]]}
            save={saveCombineImage}
            cancel={cancelCombine}
          />
        </div>
      )}
      {props.spotImagesByPage.map((images, i) => (
        <div key={i + images.length}>
          {!!images.length && (
            <>
              <h3 className="text-xl font-bold underline">Page {i + 1}</h3>
              <div className="flex w-full flex-wrap justify-start gap-2 px-2">
                {images.map((image, j) => (
                  <button
                    key={`${image.id}${image.data?.substring(20, 50)}`}
                    className={cn(
                      "flex h-fit flex-col justify-between gap-1 rounded-xl border px-2 pt-4 text-center shadow",
                      {
                        "border-neutral-400 bg-neutral-50 text-neutral-800 shadow-neutral-900/30":
                          spot1?.[0] !== image && spot2?.[0] !== image,
                        "border-green-400 bg-green-50 text-green-800 shadow-green-900/30":
                          spot1?.[0] === image || spot2?.[0] === image,
                      },
                    )}
                    onClick={() => toggleCombine(image, i, j)}
                  >
                    <figure>
                      <img
                        src={image.data}
                        alt={image.id}
                        width={image.width}
                        height={image.height}
                        className="h-auto max-w-80"
                      />
                      <figcaption className="py-1 font-medium">
                        Page {i + 1}, Selection {j + 1}
                      </figcaption>
                    </figure>
                  </button>
                ))}
              </div>
            </>
          )}
        </div>
      ))}
    </div>
  );
}

export function CombineCanvas(props: {
  combineImageData: [CroppedImageData, CroppedImageData];
  save: (data: string, height: number, width: number) => void;
  cancel: () => void;
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null!);
  const [offsetY, setOffsetY] = useState(0);

  useEffect(() => {
    const ctx = canvasRef.current.getContext("2d")!;
    const images: HTMLImageElement[] = [];
    canvasRef.current.width = 0;
    canvasRef.current.height = 0;
    // TODO: unfold this loop to run exactly twice and handle the offset
    for (const imageData of props.combineImageData) {
      if (imageData.height + Math.abs(offsetY) > canvasRef.current.height) {
        canvasRef.current.height = imageData.height + Math.abs(offsetY);
      }
      canvasRef.current.width += imageData.width;
      const img = new Image(imageData.width, imageData.height);
      img.src = imageData.data;
      images.push(img);
    }
    ctx.fillStyle = "white";
    ctx.fillRect(0, 0, canvasRef.current.width, canvasRef.current.height);
    // if offsetY is positive, the second image is being moved down, so render the first image from the top
    // and the second image down by the offset
    if (offsetY > 0) {
      // first image
      ctx.drawImage(images[0], 0, 0, images[0].width, images[0].height);
      //second image slid over by the width of the first image plus 10px padding
      ctx.drawImage(
        images[1],
        images[0].width + 10,
        offsetY,
        images[1].width,
        images[1].height,
      );
    } else {
      // if offsetY is negative, the second image is being moved up, so render the first image lowered by
      // the offset and the second image at the top
      // first image
      ctx.drawImage(images[0], 0, -offsetY, images[0].width, images[0].height);
      //second image slid over by the width of the first image plus 10px padding
      ctx.drawImage(
        images[1],
        images[0].width + 10,
        0,
        images[1].width,
        images[1].height,
      );
    }
  }, [props.combineImageData, offsetY]);

  const handleSave = useCallback(() => {
    console.log("saving");
    const data = canvasRef.current.toDataURL();
    props.save(data, canvasRef.current.height, canvasRef.current.width);
  }, [props]);

  return (
    <div className="flex max-w-full flex-col gap-2 rounded-xl border border-neutral-400 bg-white p-2 shadow-sm shadow-neutral-900/30">
      <div className="flex flex-wrap justify-end gap-2">
        <button
          className="action-button red focusable mr-4"
          onClick={props.cancel}
        >
          <span className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6" />
          Cancel
        </button>
        <button
          className="action-button purple focusable"
          onClick={() => setOffsetY(offsetY - 10)}
        >
          <span className="icon-[iconamoon--arrow-up-5-circle-thin] -ml-1 size-6" />
          Up
        </button>
        <button
          className="action-button yellow focusable mr-4"
          onClick={() => setOffsetY(offsetY + 10)}
        >
          <span className="icon-[iconamoon--arrow-down-5-circle-thin] -ml-1 size-6" />
          Down
        </button>
        <button className="action-button green" onClick={() => handleSave()}>
          <span className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-6" />
          Save
        </button>
      </div>
      <canvas ref={canvasRef} className="w-full bg-transparent" />
    </div>
  );
}

export function SaveMode(props: {
  spotImagesByPage: CroppedImageData[][];
  csrf: string;
  pieceid: string;
  goBack: () => void;
}) {
  const [spotImages, setSpotImages] = useState<CroppedImageData[]>([]);
  const formRef = useRef<HTMLFormElement>(null!);

  useEffect(() => {
    const imagesFlattened: CroppedImageData[] = [];
    for (const images of props.spotImagesByPage) {
      if (images.length) {
        imagesFlattened.push(...images);
      }
    }
    setSpotImages(imagesFlattened);
  }, [props.spotImagesByPage]);

  const handleSubmit = useCallback(
    (evt: Event) => {
      evt.preventDefault();
      // TODO: errors should show alert and try to save something with the progress
      Promise.all(
        spotImages.map((image) => fetch(image.data).then((res) => res.blob())),
      )
        .then((images) => {
          const fd = new FormData(formRef.current);
          fd.append("numSpots", `${images.length}`);
          for (let i = 0; i < images.length; i++) {
            fd.append(`spots.${i}.image`, images[i]);
          }
          fetch(`/library/pieces/${props.pieceid}/spots/pdf`, {
            method: "POST",
            body: fd,
          })
            .then((res) => {
              if (res.ok) {
                res
                  .json()
                  .then((body) => {
                    globalThis.dispatchEvent(
                      new CustomEvent("ShowAlert", {
                        detail: {
                          variant: "success",
                          title: "Added",
                          message: "Your spot(s) have been added!",
                          duration: 3000,
                        },
                      }),
                    );

                    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                    if (body?.error) {
                      globalThis.dispatchEvent(
                        new CustomEvent("ShowAlert", {
                          detail: {
                            variant: "warning",
                            title: "Error",
                            message: "Some spots could not be added.",
                            duration: 3000,
                          },
                        }),
                      );
                    }

                    return htmx.ajax(
                      "GET",
                      `/library/pieces/${props.pieceid}`,
                      "#main-content",
                    );
                  })
                  .then()
                  .catch(console.error);
              } else {
                res.text().then(alert).catch(console.log);
              }
            })
            .catch(console.error);
        })
        .catch(console.error);
    },
    [props.pieceid, spotImages],
  );

  return (
    <form
      onSubmit={handleSubmit}
      action="#"
      className="flex w-full flex-col"
      ref={formRef}
    >
      <div className="flex w-full flex-wrap items-center justify-between gap-2 py-4">
        <button
          type="button"
          className="action-button amber focusable"
          onClick={props.goBack}
        >
          <span className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6" />
          Go Back
        </button>
        <h3 className="text-3xl font-bold">Save</h3>
        <button className="action-button green focusable">
          Save
          <span
            class="icon-[iconamoon--check-circle-1-thin] -mr-1 size-6"
            aria-hidden="true"
          />
        </button>
      </div>
      <input type="hidden" name="gorilla.csrf.Token" value={props.csrf} />
      <div className="flex w-full flex-wrap justify-center gap-2 px-2">
        {spotImages.map((image, i) => (
          <div
            key={`${image.id}${image.data?.substring(20, 50)}`}
            className="flex h-fit flex-col justify-between gap-1 rounded-xl border border-neutral-400 bg-neutral-50 px-2 pt-4 text-center text-neutral-800 shadow shadow-neutral-900/30"
          >
            <figure>
              <img
                src={image.data}
                alt={image.id}
                width={image.width}
                height={image.height}
                className="h-auto max-w-80"
              />
              <figcaption className="py-1 font-medium">
                <input
                  type="text"
                  className="basic-field"
                  defaultValue={`Spot ${i + 1}`}
                  name={`spots.${i}.name`}
                />
              </figcaption>
            </figure>
          </div>
        ))}
      </div>
    </form>
  );
}
