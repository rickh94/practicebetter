// import { Suspense, lazy } from "preact/compat";
import { useCallback, useMemo, useRef, useState } from "preact/hooks";
import { AddAudioPrompt, AddImagePrompt } from "../pieces/add-prompts";
import { cn } from "../common";
import NotesDisplay from "./notes-display";

// const NotesDisplay = lazy(() => import("./notes-display"));

export function AudioPromptSummary({
  url,
  spotid,
  pieceid,
  save,
  csrf,
}: {
  url: string;
  csrf?: string;
  spotid?: string;
  pieceid?: string;
  save?: (url: string) => void;
}) {
  const [displayUrl, setDisplayUrl] = useState(url);
  const saveAudio = useCallback(
    (url: string) => {
      setDisplayUrl(url);
      save?.(url);
    },
    [save],
  );

  if (!displayUrl) {
    return (
      <div className="flex w-full items-center justify-between">
        <div>No Audio Prompt</div>
        {csrf && spotid ? (
          <AddAudioPrompt
            small
            save={saveAudio}
            audioPromptUrl=""
            csrf={csrf}
            spotid={spotid}
            pieceid={pieceid}
          />
        ) : null}
      </div>
    );
  }
  return (
    <details>
      <summary className="focusable purple flex cursor-pointer select-none items-center justify-between gap-1 rounded-xl  border border-purple-400 bg-purple-200 py-2 pl-4 pr-2 font-medium text-purple-800 shadow-sm shadow-purple-900/30 transition duration-200 hover:border-purple-500 hover:bg-purple-300 hover:shadow hover:shadow-purple-900/50">
        <div className="flex items-center gap-2">
          <span
            className="-ml icon-[iconamoon--volume-up-thin] size-5"
            aria-hidden="true"
          />
          Audio Prompt
        </div>
        <span className="summary-icon icon-[iconamoon--arrow-right-6-circle-thin] size-6 transition-transform" />
      </summary>
      <audio controls className="my-1 w-full py-1">
        <source src={displayUrl} type="audio/mpeg" />
      </audio>
    </details>
  );
}

export function ImagePromptWC({
  url,
  csrf,
  spotid,
  pieceid,
}: {
  url: string;
  csrf?: string;
  spotid?: string;
  pieceid?: string;
}) {
  return (
    <ImagePromptSummary
      url={url}
      csrf={csrf}
      spotid={spotid}
      pieceid={pieceid}
    />
  );
}

export function ImagePromptSummary({
  url,
  csrf,
  spotid,
  pieceid,
  save,
}: {
  url: string;
  csrf?: string;
  spotid?: string;
  pieceid?: string;
  save?: (url: string) => void;
}) {
  const lightboxRef = useRef<HTMLDialogElement>(null);
  const [displayUrl, setDisplayUrl] = useState(url);

  const showBig = useCallback(() => {
    if (lightboxRef.current) {
      lightboxRef.current.showModal();
      globalThis.handleShowModal();
    }
  }, []);

  const closeBig = useCallback(() => {
    globalThis.handleCloseModal();
    if (lightboxRef.current) {
      if (lightboxRef.current) {
        lightboxRef.current.classList.add("close");
        setTimeout(() => {
          lightboxRef.current?.close();
          lightboxRef.current?.classList.remove("close");
        }, 150);
      }
    }
  }, []);

  const saveImage = useCallback(
    (url: string) => {
      setDisplayUrl(url);
      save?.(url);
    },
    [save],
  );

  if (!displayUrl) {
    return (
      <div className="flex w-full items-center justify-between">
        <div>No Image Prompt</div>
        {csrf && spotid ? (
          <AddImagePrompt
            small
            save={saveImage}
            imagePromptUrl=""
            csrf={csrf}
            spotid={spotid}
            pieceid={pieceid}
          />
        ) : null}
      </div>
    );
  }
  return (
    <>
      <details>
        <summary className="focusable cyan flex cursor-pointer select-none items-center justify-between gap-1 rounded-xl  border border-cyan-400 bg-cyan-200 py-2 pl-4 pr-2 font-medium text-cyan-800 shadow-sm shadow-cyan-900/30 transition duration-200 hover:border-cyan-500 hover:bg-cyan-300 hover:shadow hover:shadow-cyan-900/50">
          <div className="flex items-center gap-2">
            <span
              className="icon-[iconamoon--file-image-thin] size-5"
              aria-hidden="true"
            />
            Image Prompt
          </div>

          <span className="summary-icon icon-[iconamoon--arrow-right-6-circle-thin] size-6 transition-transform" />
        </summary>
        <button
          onClick={showBig}
          className="focusable cyan mt-2 rounded-xl p-2"
        >
          <figure className="my-2 w-full">
            <img
              src={displayUrl}
              width={480}
              height={120}
              alt="Image Prompt"
              className="my-2 w-full"
            />
            <figcaption className="text-sm">Click to view larger</figcaption>
          </figure>
        </button>
      </details>
      <dialog ref={lightboxRef} className="p-0">
        <button
          onClick={closeBig}
          className="focusable cyan m-0 rounded-xl p-0"
        >
          <figure className="w-full sm:max-w-3xl">
            <img
              src={displayUrl}
              width={480}
              height={120}
              alt="Image Prompt"
              className="h-auto w-full"
            />
            <figcaption className="text-sm">Click to Close</figcaption>
          </figure>
        </button>
      </dialog>
    </>
  );
}

export function NotesPromptSummary({ notes }: { notes: string }) {
  if (!notes) {
    return <div>No Notes Prompt</div>;
  }
  return (
    <details>
      <summary className="focusable fuchsia flex cursor-pointer select-none items-center justify-between gap-1 rounded-xl  border border-fuchsia-400 bg-fuchsia-200 py-2 pl-4 pr-2 font-medium text-fuchsia-800 shadow-sm shadow-fuchsia-900/30 transition duration-200 hover:border-fuchsia-500 hover:bg-fuchsia-300 hover:shadow hover:shadow-fuchsia-900/50">
        <div className="flex items-center gap-2">
          <span
            className="icon-[iconamoon--music-2-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Notes Prompt
        </div>
        <span className="summary-icon icon-[iconamoon--arrow-right-6-circle-thin] size-6 transition-transform" />
      </summary>
      <div className="min-h-[6rem] w-full">
        <NotesDisplay
          staffwidth={500}
          wrap={{
            minSpacing: 0,
            maxSpacing: 0,
            preferredMeasuresPerLine: 2,
          }}
          responsive="resize"
          notes={notes}
        />
      </div>
    </details>
  );
}
// <Suspense fallback={<div className="my-2">Loading Notes...</div>}>
// </Suspense>

export function RemindersSummary({
  text,
  csrf,
  pieceid,
  spotid,
  save,
}: {
  text: string;
  pieceid?: string;
  spotid?: string;
  csrf?: string;
  save?: (text: string) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const formRef = useRef(null);
  const [displayText, setDisplayText] = useState<string>(
    text || "No Reminders",
  );

  const startEditing = useCallback(() => {
    setIsEditing(true);
  }, []);

  const stopEditing = useCallback(() => {
    setIsEditing(false);
  }, []);

  const onSubmit = useCallback(
    async (e: SubmitEvent) => {
      e.preventDefault();
      if (!formRef.current) {
        return;
      }
      const data = new FormData(formRef.current);
      try {
        const res = await fetch(
          `/library/pieces/${pieceid}/spots/${spotid}/reminders`,
          {
            method: "PATCH",
            body: data,
          },
        );
        if (res.ok) {
          if (data.get("text")) {
            save?.(data.get("text") as string);
            setDisplayText((data.get("text") as string) || "No Reminders");
          }
          stopEditing();
        } else {
          console.log(await res.text());
          globalThis.dispatchEvent(
            new CustomEvent("ShowAlert", {
              detail: {
                variant: "error",
                title: "Error",
                message: "Failed to save reminders",
                duration: 5000,
              },
            }),
          );
        }
      } catch (err) {
        console.error(err);
        globalThis.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              variant: "error",
              title: "Error",
              message: "Failed to save reminders",
              duration: 5000,
            },
          }),
        );
      }
    },
    [pieceid, save, spotid, stopEditing],
  );

  const defaultValue = useMemo(() => {
    if (displayText == "No Reminders") {
      return "";
    }
    return displayText ?? "";
  }, [displayText]);

  return (
    <details open id={`reminder-details-${pieceid}-${spotid}`}>
      <summary className="focusable blue flex cursor-pointer select-none items-center justify-between gap-1 rounded-xl  border border-blue-400 bg-blue-200 py-2 pl-4 pr-2 font-medium text-blue-800 shadow-sm shadow-blue-900/30 transition duration-200 hover:border-blue-500 hover:bg-blue-300 hover:shadow hover:shadow-blue-900/50">
        <div className="flex items-center gap-2">
          <span
            className="icon-[ph--chat-centered-text-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Reminders
        </div>
        <span className="summary-icon icon-[iconamoon--arrow-right-6-circle-thin] size-6 transition-transform" />
      </summary>

      {isEditing ? (
        <form
          className="flex flex-col gap-2 py-1 sm:flex-row"
          ref={formRef}
          onSubmit={onSubmit}
        >
          <input type="hidden" name="gorilla.csrf.Token" value={csrf} />
          <div className="flex h-max flex-grow flex-col gap-1 sm:min-h-12 sm:flex-row">
            <textarea
              name="text"
              defaultValue={defaultValue}
              placeholder="Add some reminders"
              className={cn(
                "focusable basic-field min-h-[5.25rem] w-full flex-grow",
              )}
            />
          </div>
          <div className="flex min-h-20 flex-grow-0 flex-col justify-center gap-1">
            <button className="focusable action-button blue" type="submit">
              <span
                className="icon-[iconamoon--arrow-up-5-circle-thin] -ml-1 size-6"
                aria-hidden="true"
              />
              Save
            </button>
            <button
              onClick={stopEditing}
              className="focusable action-button red"
            >
              <span
                className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
                aria-hidden="true"
              />
              Cancel
            </button>
          </div>
        </form>
      ) : (
        <div
          className="flex flex-col py-1 sm:min-h-12 sm:flex-row"
          id="reminder-details"
        >
          <p className="min-h-12 flex-grow py-1 font-semibold">
            {displayText ?? null}
          </p>
          <button
            className="focusable action-button blue px-6 py-2"
            onClick={startEditing}
          >
            <span
              className="icon-[iconamoon--edit-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Edit
          </button>
        </div>
      )}
    </details>
  );
}
