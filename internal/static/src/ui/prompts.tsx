import {
  ChevronRightIcon,
  MusicalNoteIcon,
  XMarkIcon,
  ChatBubbleBottomCenterTextIcon,
  PencilIcon,
  PhotoIcon,
  SpeakerWaveIcon,
} from "@heroicons/react/24/solid";
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
      <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-yellow-500/50 py-2 pl-4 pr-2 font-semibold text-yellow-800 transition duration-200 hover:bg-yellow-300/50">
        <div className="flex items-center gap-2">
          <SpeakerWaveIcon className="-ml-1 size-5" />
          Audio Prompt
        </div>
        <ChevronRightIcon className="summary-icon size-6 transition-transform" />
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
    lightboxRef.current?.showModal();
  }, [lightboxRef.current]);

  const closeBig = useCallback(() => {
    if (lightboxRef.current) {
      lightboxRef.current.classList.add("close");
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          if (lightboxRef.current) {
            lightboxRef.current.classList.remove("close");
            lightboxRef.current.close();
          }
        });
      });
    }
  }, [lightboxRef.current]);

  if (!url) {
    return <div>No Image Prompt</div>;
  }
  return (
    <>
      <details>
        <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-indigo-500/50 py-2 pl-4 pr-2 font-semibold text-indigo-800 transition duration-200 hover:bg-indigo-300/50">
          <div className="flex items-center gap-2">
            <PhotoIcon className="-ml-1 size-5" />
            Image Prompt
          </div>
          <ChevronRightIcon className="summary-icon size-6 transition-transform" />
        </summary>
        <button onClick={showBig} className="m-0 p-0">
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
        <button onClick={closeBig} className="m-0 p-0">
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
      <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-sky-500/50 py-2 pl-4 pr-2 font-semibold text-sky-800 transition duration-200 hover:bg-sky-300/50">
        <div className="flex items-center gap-2">
          <MusicalNoteIcon className="-ml-1 size-5" />
          Notes Prompt
        </div>
        <ChevronRightIcon className="summary-icon size-6 transition-transform" />
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
      <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-lime-500/50 py-2 pl-4 pr-2 font-semibold text-lime-800 transition duration-200 hover:bg-lime-300/50">
        <div className="flex items-center gap-2">
          <ChatBubbleBottomCenterTextIcon className="-ml-1 size-5" />
          Reminders
        </div>
        <ChevronRightIcon className="summary-icon size-6 transition-transform" />
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
            className="focusable flex items-center justify-center gap-1 rounded-xl bg-lime-700/10 px-6 py-2 font-semibold text-lime-800 transition duration-200 hover:bg-lime-700/20"
            href={`/library/pieces/${pieceid}/spots/${spotid}/reminders/edit`}
            target={`#${targetId}`}
          >
            <PencilIcon className="-ml-1 size-5" />
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
      <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-lime-500/50 py-2 pl-4 pr-2 font-semibold text-lime-800 transition duration-200 hover:bg-lime-300/50">
        <div className="flex items-center gap-2">
          <ChatBubbleBottomCenterTextIcon className="-ml-1 size-5" />
          Reminders
        </div>
        <ChevronRightIcon className="summary-icon size-6 transition-transform" />
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
              "focusable min-h-[5.25rem] w-full flex-grow rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20",
            )}
          />
          {!!error && <p className="italic text-red-600">{error}</p>}
        </div>
        <div className="flex min-h-20 flex-grow-0 flex-col justify-center gap-1">
          <button
            className="focusable flex items-center justify-center gap-1 rounded-xl bg-lime-700/10 px-6 py-2 font-semibold text-lime-800 transition duration-200 hover:bg-lime-700/20"
            type="submit"
          >
            <PencilIcon className="-ml-1 size-5" />
            Save
          </button>
          <Link
            pushUrl={false}
            className="focusable flex items-center justify-center gap-1 rounded-xl bg-red-700/10 px-6 py-2 font-semibold text-red-800 transition duration-200 hover:bg-red-700/20"
            href={`/library/pieces/${pieceid}/spots/${spotid}/reminders`}
            target={`#${id}`}
          >
            <XMarkIcon className="-ml-1 size-5" />
            Cancel
          </Link>
        </div>
      </form>
    </details>
  );
}
