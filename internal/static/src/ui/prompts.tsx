import { Suspense, lazy } from "preact/compat";
import { useCallback, useRef } from "preact/hooks";
import { Link } from "./links";
import { cn } from "../common";

const NotesDisplay = lazy(() => import("./notes-display"));

export function AudioPromptSummary({ url }: { url: string }) {
  if (!url) {
    return <div>No Audio Prompt</div>;
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
        <source src={url} type="audio/mpeg" />
      </audio>
    </details>
  );
}
export function ImagePromptSummary({ url }: { url: string }) {
  const lightboxRef = useRef<HTMLDialogElement>(null);

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

  if (!url) {
    return <div>No Image Prompt</div>;
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
              src={url}
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
              src={url}
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
        <Suspense fallback={<div className="my-2">Loading Notes...</div>}>
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
        </Suspense>
      </div>
    </details>
  );
}

export function RemindersSummary({
  text,
  pieceid = "",
  spotid = "",
  id = "",
}: {
  text: string;
  pieceid?: string;
  spotid?: string;
  id?: string;
}) {
  let targetId: string;
  if (id) {
    targetId = id;
  } else {
    targetId = `reminder-details-${pieceid}-${spotid}`;
  }
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
      <div
        className="flex flex-col py-1 sm:min-h-12 sm:flex-row"
        id="reminder-details"
      >
        <p className="min-h-12 flex-grow py-1 font-semibold">
          {text?.length > 0 ? text : "No Reminders"}
        </p>
        {!!pieceid && !!spotid && (
          <Link
            pushUrl={false}
            className="focusable action-button blue px-6 py-2"
            href={`/library/pieces/${pieceid}/spots/${spotid}/reminders/edit`}
            target={`#${targetId}`}
          >
            <span
              className="icon-[iconamoon--edit-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Edit
          </Link>
        )}
      </div>
    </details>
  );
}

export function EditRemindersSummary({
  text,
  pieceid = "",
  spotid = "",
  id = "",
  csrf,
  error = "",
}: {
  text: string;
  pieceid: string;
  spotid: string;
  id?: string;
  csrf: string;
  error?: string;
}) {
  return (
    <details open>
      <summary className="focusable blue flex cursor-pointer select-none items-center justify-between gap-1  rounded-xl border border-blue-400 bg-blue-200 py-2 pl-4 pr-2 font-medium text-blue-800 shadow-sm shadow-blue-900/30 transition duration-200 hover:border-blue-500 hover:bg-blue-300 hover:shadow hover:shadow-blue-900/50">
        <div className="flex items-center gap-2">
          <span
            className="icon-[ph--chat-centered-text-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Reminders
        </div>
        <span className="summary-icon icon-[iconamoon--arrow-right-6-circle-thin] size-6 transition-transform" />
      </summary>
      <form
        className="flex flex-col gap-2 py-1 sm:flex-row"
        id="reminder-details-form"
        hx-target={`#${id}`}
        hx-post={`/library/pieces/${pieceid}/spots/${spotid}/reminders`}
        hx-swap="outerHTML transition:true"
        hx-push-url="false"
      >
        <input type="hidden" name="gorilla.csrf.Token" value={csrf} />
        <div className="flex h-max flex-grow flex-col gap-1 sm:min-h-12 sm:flex-row">
          <textarea
            name="text"
            defaultValue={text ?? ""}
            placeholder="Add some reminders"
            className={cn(
              "focusable basic-field min-h-[5.25rem] w-full flex-grow",
            )}
          />
          {!!error && <p className="italic text-red-600">{error}</p>}
        </div>
        <div className="flex min-h-20 flex-grow-0 flex-col justify-center gap-1">
          <button className="focusable action-button blue" type="submit">
            <span
              className="icon-[iconamoon--arrow-up-5-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Save
          </button>
          <Link
            pushUrl={false}
            className="focusable action-button red"
            href={`/library/pieces/${pieceid}/spots/${spotid}/reminders`}
            target={`#${id}`}
          >
            <span
              className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Cancel
          </Link>
        </div>
      </form>
    </details>
  );
}
