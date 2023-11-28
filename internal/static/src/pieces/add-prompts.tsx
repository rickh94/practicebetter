import {
  CheckIcon,
  SpeakerWaveIcon,
  ArrowPathIcon,
  PhotoIcon,
  DocumentTextIcon,
  MusicalNoteIcon,
} from "@heroicons/react/20/solid";
import { lazy, Suspense } from "preact/compat";
import { useState, useCallback, useRef } from "preact/hooks";
import { UseFormRegisterReturn } from "react-hook-form";
import { cn } from "../common";
import { ColorlessButton, HappyButton } from "../ui/buttons";
// import NotesDisplay from "../ui/notes-display";
const NotesDisplay = lazy(() => import("../ui/notes-display"));

// TODO: change these to my native dialogs
// TODO: figure out file uploading

export function AddAudioPrompt({
  audioPromptUrl,
  registerReturn,
  save,
}: {
  audioPromptUrl?: string | null;
  registerReturn: UseFormRegisterReturn;
  save: (url: string) => void;
}) {
  const [isUploading, setIsUploading] = useState(false);
  const ref = useRef<HTMLDialogElement>(null);

  const open = useCallback(
    function () {
      if (ref.current) {
        ref.current.showModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      if (ref.current) {
        ref.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (ref.current) {
              ref.current.classList.remove("close");
              ref.current.close();
            }
          });
        });
      }
    },
    [ref],
  );

  return (
    <>
      <ColorlessButton
        onClick={open}
        className={cn(
          audioPromptUrl ? "bg-yellow-500/50" : "bg-yellow-700/10",
          "text-yellow-800",
        )}
      >
        {audioPromptUrl ? (
          <>
            <span className="sr-only">Checked</span>
            <CheckIcon className="h-4 w-4" />
          </>
        ) : (
          <SpeakerWaveIcon className="h-4 w-4" />
        )}
        Audio
      </ColorlessButton>
      <dialog
        ref={ref}
        aria-labelledby="add-audio-prompt-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="add-audio-prompt-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Add Audio Prompt
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          Upload an audio file (max 512KB) or paste in a public URL to audio
          that will prompt you for this spot.
        </div>
        <div className="flex w-full flex-col">
          <label
            className="text-left text-sm font-medium leading-6 text-neutral-900"
            htmlFor="url"
          >
            Url
          </label>
          <div className="flex items-center gap-0">
            <input
              id="url"
              {...registerReturn}
              className="focusable rounded-r-0 w-full rounded-l-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
          </div>
        </div>
        <HappyButton
          grow
          disabled={isUploading}
          onClick={close}
          className="mt-4 w-full"
        >
          {isUploading ? (
            <ArrowPathIcon className="h-6 w-6" />
          ) : (
            <CheckIcon className="h-6 w-6" />
          )}
          {isUploading ? "Please Wait..." : "Done"}
        </HappyButton>
      </dialog>
    </>
  );
}

export function AddImagePrompt({
  imagePromptUrl,
  registerReturn,
  save,
}: {
  imagePromptUrl?: string | null;
  registerReturn: UseFormRegisterReturn;
  save: (url: string) => void;
}) {
  const [isUploading, setIsUploading] = useState(false);
  const ref = useRef<HTMLDialogElement>(null);

  const open = useCallback(
    function () {
      if (ref.current) {
        ref.current.showModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      if (ref.current) {
        ref.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (ref.current) {
              ref.current.classList.remove("close");
              ref.current.close();
            }
          });
        });
      }
    },
    [ref.current],
  );

  return (
    <>
      <ColorlessButton
        onClick={open}
        className={cn(
          imagePromptUrl ? "bg-indigo-500/50" : "bg-indigo-700/10",
          "text-indigo-800",
        )}
      >
        {imagePromptUrl ? (
          <>
            <span className="sr-only">Checked</span>
            <CheckIcon className="h-4 w-4" />
          </>
        ) : (
          <PhotoIcon className="h-4 w-4" />
        )}
        Image
      </ColorlessButton>

      <dialog
        ref={ref}
        aria-labelledby="add-imate-prompt-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="add-imate-prompt-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Add Image Prompt
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          Upload an image or screenshot (max 512KB) or enter a public URL for an
          image to use as a prompt for this spot.
        </div>
        <div className="flex w-full flex-col">
          <label
            className="text-left text-sm font-medium leading-6 text-neutral-900"
            htmlFor="url"
          >
            Url
          </label>
          <div className="flex items-center gap-0">
            <input
              {...registerReturn}
              id="url"
              className="focusable rounded-r-0 w-full rounded-l-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
          </div>
        </div>
        <HappyButton
          grow
          disabled={isUploading}
          onClick={close}
          className="mt-4 w-full"
        >
          {isUploading ? (
            <ArrowPathIcon className="h-6 w-6" />
          ) : (
            <CheckIcon className="h-6 w-6" />
          )}
          {isUploading ? "Please Wait..." : "Done"}
        </HappyButton>
      </dialog>
    </>
  );
}

export function AddTextPrompt({
  textPrompt,
  registerReturn,
}: {
  textPrompt?: string | null;
  registerReturn: UseFormRegisterReturn;
}) {
  const ref = useRef<HTMLDialogElement>(null);

  const open = useCallback(
    function () {
      if (ref.current) {
        ref.current.showModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      if (ref.current) {
        ref.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (ref.current) {
              ref.current.classList.remove("close");
              ref.current.close();
            }
          });
        });
      }
    },
    [ref],
  );

  return (
    <>
      <ColorlessButton
        onClick={open}
        className={cn(
          textPrompt ? "bg-lime-500/50" : "bg-lime-700/10",
          "text-lime-800",
        )}
      >
        {!!textPrompt ? (
          <>
            <span className="sr-only">Checked</span>
            <CheckIcon className="h-4 w-4" />
          </>
        ) : (
          <DocumentTextIcon className="h-4 w-4" />
        )}
        Text
      </ColorlessButton>
      <dialog
        ref={ref}
        aria-labelledby="add-text-prompt-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="add-text-prompt-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Add Text Prompt
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          Enter some text to remind yourself about this spot.
        </div>
        <div className="flex w-full flex-col">
          <label
            className="text-left text-sm font-medium leading-6 text-neutral-900"
            htmlFor="textPrompt"
          >
            Text Prompt
          </label>
          <textarea
            {...registerReturn}
            id="text"
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
          />
        </div>
        <HappyButton grow onClick={close} className="mt-4 w-full">
          <CheckIcon className="h-6 w-6" />
          Done
        </HappyButton>
      </dialog>
    </>
  );
}

export function AddNotesPrompt({
  notesPrompt,
  registerReturn,
}: {
  notesPrompt?: string | null;
  registerReturn: UseFormRegisterReturn;
}) {
  const ref = useRef<HTMLDialogElement>(null);

  const open = useCallback(
    function () {
      if (ref.current) {
        ref.current.showModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      if (ref.current) {
        ref.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (ref.current) {
              ref.current.classList.remove("close");
              ref.current.close();
            }
          });
        });
      }
    },
    [ref],
  );

  return (
    <>
      <ColorlessButton
        type="button"
        onClick={open}
        className={cn(
          notesPrompt ? "bg-sky-500/50" : "bg-sky-700/10",
          "text-sky-500",
        )}
      >
        {notesPrompt ? (
          <>
            <span className="sr-only">Checked</span>
            <CheckIcon className="h-4 w-4" />
          </>
        ) : (
          <MusicalNoteIcon className="h-4 w-4" />
        )}
        Notes
      </ColorlessButton>
      <dialog
        ref={ref}
        aria-labelledby="add-notes-prompt-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="add-notes-prompt-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Add Notes Prompt
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          The text box below will treated as{" "}
          <a
            href="https://abcnotation.com/"
            className="underline"
            target="_blank"
            rel="noopener noreferrer"
          >
            ABC notation
          </a>{" "}
          and rendered below.{" "}
        </div>
        <div className="flex w-full flex-col">
          <label
            className="text-left text-sm font-medium leading-6 text-neutral-900"
            htmlFor="notes"
          >
            ABC Notes
          </label>
          <textarea
            {...registerReturn}
            id="notes"
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            placeholder="Enter Notes using ABC notation"
          />
        </div>
        <div className="h-[6rem] w-full">
          <Suspense fallback={<div>Loading notes...</div>}>
            <NotesDisplay
              staffwidth={500}
              wrap={{
                minSpacing: 0,
                maxSpacing: 0,
                preferredMeasuresPerLine: 2,
              }}
              notes={notesPrompt ?? ""}
              responsive="resize"
            />
          </Suspense>
        </div>
        <HappyButton grow onClick={close} className="mt-4 w-full">
          <CheckIcon className="h-6 w-6" />
          Done
        </HappyButton>
      </dialog>
    </>
  );
}
