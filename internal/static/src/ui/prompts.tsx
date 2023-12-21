import { ChevronRightIcon, MusicalNoteIcon } from "@heroicons/react/20/solid";
import {
  DocumentTextIcon,
  PhotoIcon,
  SpeakerWaveIcon,
} from "@heroicons/react/24/solid";
import { Suspense, lazy } from "preact/compat";

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
        <ChevronRightIcon className="summary-icon -mr-1 size-6 transition-transform" />
      </summary>
      <audio controls className="my-1 w-full py-1">
        <source src={url} type="audio/mpeg" />
      </audio>
    </details>
  );
}
export function ImagePromptSummary({ url }: { url: string }) {
  if (!url) {
    return <div>No Image Prompt</div>;
  }
  return (
    <details>
      <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-indigo-500/50 py-2 pl-4 pr-2 font-semibold text-indigo-800 transition duration-200 hover:bg-indigo-300/50">
        <div className="flex items-center gap-2">
          <PhotoIcon className="-ml-1 size-5" />
          Image Prompt
        </div>
        <ChevronRightIcon className="summary-icon -mr-1 size-6 transition-transform" />
      </summary>
      <img
        src={url}
        width={480}
        height={120}
        alt="Image Prompt"
        className="my-2 w-full"
      />
    </details>
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
        <ChevronRightIcon className="summary-icon -mr-1 size-6 transition-transform" />
      </summary>
      <div className="h-[6rem] w-full">
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
export function TextPromptSummary({ text }: { text: string }) {
  if (!text) {
    return <div>No Text Prompt</div>;
  }
  return (
    <details>
      <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-lime-500/50 py-2 pl-4 pr-2 font-semibold text-lime-800 transition duration-200 hover:bg-lime-300/50">
        <div className="flex items-center gap-2">
          <DocumentTextIcon className="-ml-1 size-5" />
          Text Prompt
        </div>
        <ChevronRightIcon className="summary-icon -mr-1 size-6 transition-transform" />
      </summary>
      <p className="py-1">{text}</p>
    </details>
  );
}
