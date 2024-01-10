import { lazy, Suspense } from "preact/compat";
import { useState, useCallback, useRef } from "preact/hooks";
import { UseFormRegisterReturn } from "react-hook-form";
import { cn } from "../common";
import {
  AngryButton,
  ColorlessButton,
  HappyButton,
  WarningButton,
} from "../ui/buttons";
const NotesDisplay = lazy(() => import("../ui/notes-display"));

// TODO: this might be possible without custom elements

export function AddAudioPrompt({
  audioPromptUrl,
  csrf,
  save,
}: {
  audioPromptUrl?: string | null;
  csrf: string;
  save: (url: string) => void;
}) {
  const [isUploading, setIsUploading] = useState(false);
  const ref = useRef<HTMLDialogElement>(null);
  const formRef = useRef<HTMLFormElement>(null);

  const open = useCallback(
    function () {
      if (ref.current) {
        ref.current.showModal();
        globalThis.handleOpenModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      globalThis.handleCloseModal();
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

  const handleSubmit = useCallback(
    async function (e: Event) {
      e.preventDefault();
      setIsUploading(true);
      if (!formRef.current) {
        return;
      }
      const formData = new FormData(formRef.current);
      const res = await fetch("/library/upload/audio", {
        method: "POST",
        body: formData,
      });
      if (res.ok) {
        const { filename, url } = await res.json();
        setIsUploading(false);
        formRef.current.reset();

        document.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              message: `${filename} has been uploaded successfully!`,
              title: "Upload Complete",
              variant: "success",
              duration: 3000,
            },
          }),
        );

        save(url);
        close();
      } else {
        setIsUploading(false);
        document.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              message: "Upload failed!",
              title: "Upload Failed",
              variant: "error",
              duration: 3000,
            },
          }),
        );
      }
    },
    [formRef.current, setIsUploading, save, close],
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
            <span
              className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-5"
              aria-hidden="true"
            />
          </>
        ) : (
          <span
            className="-ml icon-[iconamoon--volume-up-thin] size-5"
            aria-hidden="true"
          />
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
        {audioPromptUrl ? (
          <>
            <div className="w-full text-left text-lg font-medium">
              Current File is:{" "}
              <strong className="font-bold">
                {audioPromptUrl.split("/").pop()}
              </strong>{" "}
              <span className="text-base font-normal">
                (may have been renamed)
              </span>
            </div>
            <div className="w-full">
              Remove this file first to upload a different one
            </div>
            <div class="mt-4 flex w-full gap-2">
              <WarningButton grow type="button" onClick={close}>
                <span
                  className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6"
                  aria-hidden="true"
                ></span>
                Close
              </WarningButton>
              <AngryButton grow type="button" onClick={() => save("")}>
                <span
                  className="icon-[iconamoon--trash-thin] -ml-1 size-6"
                  aria-hidden="true"
                ></span>
                Remove File
              </AngryButton>
            </div>
          </>
        ) : (
          <>
            <div className="prose prose-sm prose-neutral mt-2 text-left">
              Upload an audio file (max 1MB) that will prompt you for this spot.
            </div>
            <form
              className="flex w-full flex-col"
              ref={formRef}
              action="#"
              enctype="multipart/form-data"
              // @ts-ignore
              onSubmit={(e) => e.preventDefault()}
            >
              <input type="hidden" name="gorilla.csrf.Token" value={csrf} />
              <input
                type="file"
                name="file"
                accept="audio/mpeg"
                class="py-4"
                required
              />
            </form>
            <div class="mt-4 flex w-full gap-2">
              <WarningButton
                grow
                disabled={isUploading}
                type="button"
                onClick={close}
              >
                <span
                  className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6"
                  aria-hidden="true"
                ></span>
                Close
              </WarningButton>
              <HappyButton
                grow
                disabled={isUploading}
                type="button"
                onClick={handleSubmit}
              >
                {isUploading ? (
                  <span
                    className="icon-[ph--arrows-clockwise-thin] -ml-1 size-6"
                    aria-hidden="true"
                  ></span>
                ) : (
                  <span
                    className="icon-[iconamoon--cloud-upload-thin] -ml-1 size-6"
                    aria-hidden="true"
                  ></span>
                )}
                {isUploading ? "Please Wait..." : "Upload"}
              </HappyButton>
            </div>
          </>
        )}
      </dialog>
    </>
  );
}

export function AddImagePrompt({
  imagePromptUrl,
  save,
  csrf,
}: {
  csrf: string;
  imagePromptUrl?: string | null;
  save: (url: string) => void;
}) {
  const [isUploading, setIsUploading] = useState(false);
  const ref = useRef<HTMLDialogElement>(null);
  const formRef = useRef<HTMLFormElement>(null);

  const open = useCallback(
    function () {
      if (ref.current) {
        ref.current.showModal();
        globalThis.handleOpenModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      globalThis.handleCloseModal();
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

  const handleSubmit = useCallback(
    async function (e: Event) {
      e.preventDefault();
      setIsUploading(true);
      if (!formRef.current) {
        return;
      }
      const formData = new FormData(formRef.current);
      const res = await fetch("/library/upload/images", {
        method: "POST",
        body: formData,
      });
      if (res.ok) {
        const { filename, url } = await res.json();
        setIsUploading(false);
        formRef.current.reset();

        document.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              message: `${filename} has been uploaded successfully!`,
              title: "Upload Complete",
              variant: "success",
              duration: 3000,
            },
          }),
        );

        save(url);
        close();
      } else {
        setIsUploading(false);
        document.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              message: "Failed to upload image",
              title: "Upload Failed",
              variant: "error",
              duration: 3000,
            },
          }),
        );
      }
    },
    [formRef.current, setIsUploading, save, close],
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
            <span
              className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>
          </>
        ) : (
          <span
            className="icon-[iconamoon--file-image-thin] size-5"
            aria-hidden="true"
          ></span>
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
        {imagePromptUrl ? (
          <>
            <div className="w-full text-left text-lg font-medium">
              Current File is:{" "}
              <strong className="font-bold">
                {imagePromptUrl.split("/").pop()}
              </strong>{" "}
              <span className="text-base font-normal">
                (may have been renamed)
              </span>
            </div>
            <div className="w-full">
              Remove this file first to upload a different one
            </div>
            <div class="mt-4 flex w-full gap-2">
              <WarningButton grow type="button" onClick={close}>
                <span
                  className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6"
                  aria-hidden="true"
                ></span>
                Close
              </WarningButton>
              <AngryButton grow type="button" onClick={() => save("")}>
                <span
                  className="icon-[iconamoon--trash-thin] -ml-1 size-6"
                  aria-hidden="true"
                ></span>
                Remove File
              </AngryButton>
            </div>
          </>
        ) : (
          <>
            <div className="prose prose-sm prose-neutral mt-2 text-left">
              Upload an image file (max 1MB) that will prompt you for this spot.
              You can take a screenshot or a picture of your music with your
              phone.
            </div>
            <form
              className="flex w-full flex-col"
              ref={formRef}
              action="#"
              enctype="multipart/form-data"
              // @ts-ignore
              onSubmit={(e) => e.preventDefault()}
            >
              <input type="hidden" name="gorilla.csrf.Token" value={csrf} />
              <input
                type="file"
                name="file"
                accept="image/png,image/jpg,image/jpeg,image/gif"
                class="py-4"
              />
            </form>
            <div class="mt-4 flex w-full gap-2">
              <WarningButton
                grow
                disabled={isUploading}
                type="button"
                onClick={close}
              >
                <span
                  className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6"
                  aria-hidden="true"
                ></span>
                Close
              </WarningButton>
              <HappyButton
                grow
                disabled={isUploading}
                type="button"
                onClick={handleSubmit}
              >
                {isUploading ? (
                  <span
                    className="icon-[ph--arrows-clockwise] -ml-1 size-6"
                    aria-hidden="true"
                  ></span>
                ) : (
                  <span
                    className="icon-[iconamoon--cloud-upload-thin] -ml-1 size-6"
                    aria-hidden="true"
                  ></span>
                )}
                {isUploading ? "Please Wait..." : "Upload"}
              </HappyButton>
            </div>
          </>
        )}
      </dialog>
    </>
  );
}

export function AddReminders({
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
        globalThis.handleOpenModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      globalThis.handleCloseModal();
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
          textPrompt ? "bg-lime-500/50" : "bg-lime-700/10",
          "text-lime-800",
        )}
      >
        {!!textPrompt ? (
          <>
            <span className="sr-only">Checked</span>
            <span
              className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>
          </>
        ) : (
          <span
            className="icon-[ph--chat-centered-text-thin] -ml-1 size-5"
            aria-hidden="true"
          ></span>
        )}
        Reminders
      </ColorlessButton>
      <dialog
        ref={ref}
        aria-labelledby="add-reminders-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="add-reminders-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Add Reminders
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
            Reminders
          </label>
          <textarea
            {...registerReturn}
            id="textPrompt"
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
          />
        </div>
        <HappyButton grow onClick={close} className="mt-4 w-full">
          <span
            className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-5"
            aria-hidden="true"
          ></span>
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
        globalThis.handleOpenModal();
      }
    },
    [ref.current],
  );

  const close = useCallback(
    function () {
      globalThis.handleCloseModal();
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
          "text-sky-800",
        )}
      >
        {notesPrompt ? (
          <>
            <span className="sr-only">Checked</span>
            <span
              className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>
          </>
        ) : (
          <span
            className="icon-[iconamoon--music-2-thin] -ml-1 size-5"
            aria-hidden="true"
          ></span>
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
            rows={4}
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            placeholder="Enter Notes using ABC notation"
          />
        </div>
        <div className="w-full">
          <Suspense fallback={<div>Loading notes...</div>}>
            <NotesDisplay
              staffwidth={500}
              wrap={{
                minSpacing: 0,
                maxSpacing: 0,
                preferredMeasuresPerLine: 2,
              }}
              responsive="resize"
              notes={notesPrompt ?? ""}
            />
          </Suspense>
        </div>
        <HappyButton grow onClick={close} className="mt-4 w-full">
          <span
            className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-6"
            aria-hidden="true"
          ></span>
          Done
        </HappyButton>
      </dialog>
    </>
  );
}
