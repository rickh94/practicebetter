import { Dialog, Transition } from "@headlessui/react";
import {
  CheckIcon,
  SpeakerWaveIcon,
  ArrowPathIcon,
  PhotoIcon,
  DocumentTextIcon,
  MusicalNoteIcon,
} from "@heroicons/react/20/solid";
import { Suspense } from "preact/compat";
import { useState, useCallback } from "preact/hooks";
import { Fragment } from "preact/jsx-runtime";
import { UseFormRegisterReturn } from "react-hook-form";
import { cn } from "../common";
import { ColorlessButton, HappyButton } from "../ui/buttons";
import NotesDisplay from "../ui/notes-display";

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
  const [isOpen, setIsOpen] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  const open = useCallback(
    function () {
      setIsOpen(true);
    },
    [setIsOpen],
  );

  const close = useCallback(
    function () {
      setIsOpen(false);
    },
    [setIsOpen],
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
      <Transition.Root show={isOpen} as={Fragment}>
        {/*
        // @ts-ignore */}
        <Dialog as="div" className="relative z-10" onClose={close}>
          {/*
        // @ts-ignore */}
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="back fixed inset-0 bg-neutral-800/30 backdrop-blur-sm transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel
                  className={`relative transform overflow-hidden rounded-xl bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 pb-4 pt-5 text-left shadow-lg transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6`}
                >
                  <div>
                    <div className="mt-3 text-center sm:mt-5">
                      <Dialog.Title
                        as="h3"
                        // @ts-ignore
                        className="text-2xl font-semibold leading-6 text-neutral-900"
                      >
                        Add Audio Prompt
                      </Dialog.Title>
                      <div className="prose prose-sm prose-neutral mt-2 text-left">
                        Upload an audio file (max 512KB) or paste in a public
                        URL to audio that will prompt you for this spot.
                      </div>
                      <div className="flex flex-col">
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
                    </div>
                  </div>
                  <div className="mt-5 flex gap-4 sm:mt-6">
                    <HappyButton grow disabled={isUploading} onClick={close}>
                      {isUploading ? (
                        <ArrowPathIcon className="h-6 w-6" />
                      ) : (
                        <CheckIcon className="h-6 w-6" />
                      )}
                      {isUploading ? "Please Wait..." : "Done"}
                    </HappyButton>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
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
  const [isOpen, setIsOpen] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  const open = useCallback(
    function () {
      setIsOpen(true);
    },
    [setIsOpen],
  );

  const close = useCallback(
    function () {
      setIsOpen(false);
    },
    [setIsOpen],
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
      <Transition.Root show={isOpen} as={Fragment}>
        <Dialog
          as="div"
          // @ts-ignore
          className="relative z-10"
          onClose={close}
        >
          {/*
          // @ts-ignore */}
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="back fixed inset-0 bg-neutral-800/30 backdrop-blur-sm transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel
                  className={`relative transform overflow-hidden rounded-xl bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 pb-4 pt-5 text-left shadow-lg transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6`}
                >
                  <div>
                    <div className="mt-3 text-center sm:mt-5">
                      <Dialog.Title
                        as="h3"
                        // @ts-ignore
                        className="text-2xl font-semibold leading-6 text-neutral-900"
                      >
                        Add Image Prompt
                      </Dialog.Title>
                      <div className="prose prose-sm prose-neutral mt-2 text-left">
                        Upload an image or screenshot (max 512KB) or enter a
                        public URL for an image to use as a prompt for this
                        spot.
                      </div>
                      <div className="flex flex-col">
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
                    </div>
                  </div>
                  <div className="mt-5 flex gap-4 sm:mt-6">
                    <HappyButton grow disabled={isUploading} onClick={close}>
                      {isUploading ? (
                        <ArrowPathIcon className="h-6 w-6" />
                      ) : (
                        <CheckIcon className="h-6 w-6" />
                      )}
                      {isUploading ? "Please Wait..." : "Done"}
                    </HappyButton>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
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
  const [isOpen, setIsOpen] = useState(false);

  const open = useCallback(() => setIsOpen(true), [setIsOpen]);
  const close = useCallback(() => setIsOpen(false), [setIsOpen]);

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
      <Transition.Root show={isOpen} as={Fragment}>
        <Dialog
          as="div"
          // @ts-ignore
          className="relative z-10"
          onClose={close}
        >
          {/*
          // @ts-ignore */}
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="back fixed inset-0 bg-neutral-800/30 backdrop-blur-sm transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel
                  className={`relative transform overflow-hidden rounded-xl bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 pb-4 pt-5 text-left shadow-lg transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6`}
                >
                  <div>
                    <div className="mt-3 text-center sm:mt-5">
                      <Dialog.Title
                        as="h3"
                        // @ts-ignore
                        className="text-2xl font-semibold leading-6 text-neutral-900"
                      >
                        Add Text Prompt
                      </Dialog.Title>
                      <div className="prose prose-sm prose-neutral mt-2 text-left">
                        Enter some text to remind yourself about this spot.
                      </div>
                      <div className="flex flex-col">
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
                    </div>
                  </div>
                  <div className="mt-5 flex gap-4 sm:mt-6">
                    <HappyButton grow onClick={close}>
                      <CheckIcon className="h-6 w-6" />
                      Done
                    </HappyButton>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
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
  const [isOpen, setIsOpen] = useState(false);

  const open = useCallback(() => setIsOpen(true), [setIsOpen]);
  const close = useCallback(() => setIsOpen(false), [setIsOpen]);

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
      <Transition.Root show={isOpen} as={Fragment}>
        <Dialog
          as="div"
          // @ts-ignore
          className="relative z-10"
          onClose={close}
        >
          {/*
          // @ts-ignore */}
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="back fixed inset-0 bg-neutral-800/30 backdrop-blur-sm transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel
                  className={`relative transform overflow-hidden rounded-xl bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 pb-4 pt-5 text-left shadow-lg transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6`}
                >
                  <div>
                    <div className="mt-3 text-center sm:mt-5">
                      <Dialog.Title
                        as="h3"
                        // @ts-ignore
                        className="text-2xl font-semibold leading-6 text-neutral-900"
                      >
                        Add Notes Prompt
                      </Dialog.Title>
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
                      <div className="flex flex-col">
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
                      <div className="h-[100px]">
                        <Suspense fallback={<div>Loading notes...</div>}>
                          <NotesDisplay
                            notes={notesPrompt ?? ""}
                            responsive="resize"
                          />
                        </Suspense>
                      </div>
                    </div>
                  </div>
                  <div className="mt-5 flex gap-4 sm:mt-6">
                    <HappyButton grow onClick={close}>
                      <CheckIcon className="h-6 w-6" />
                      Done
                    </HappyButton>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
    </>
  );
}
