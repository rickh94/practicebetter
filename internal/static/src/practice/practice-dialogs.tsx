import { type Ref, useCallback } from "preact/hooks";
import { StartPracticingInterleave } from "../ui/plan-components";
import { BackToPlan } from "../ui/links";

export function ResumeDialog({
  dialogRef,
  onResume,
}: {
  dialogRef: Ref<HTMLDialogElement>;
  onResume: () => void;
}) {
  const closeDialog = useCallback(() => {
    if (dialogRef.current) {
      globalThis.handleCloseModal();
      if (dialogRef.current) {
        dialogRef.current.classList.add("close");
        setTimeout(() => {
          if (dialogRef.current) {
            dialogRef.current.classList.remove("close");
            dialogRef.current.close();
          }
        }, 150);
      }
    }
  }, [dialogRef]);

  const handleResume = useCallback(() => {
    onResume();
    closeDialog();
  }, [onResume, closeDialog]);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby="resume-title"
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
    >
      <header className="mt-2 text-center sm:text-left">
        <h3
          id="resume-title"
          className="text-2xl font-semibold leading-6 text-neutral-900"
        >
          Resume Practicing
        </h3>
      </header>
      <div className="prose prose-sm prose-neutral mt-2 text-left">
        It looks like your practicing was interrupted, would you like to resume
        it?
      </div>
      <div className="mt-2 flex w-full flex-row-reverse flex-wrap gap-2 sm:gap-2">
        <button
          onClick={handleResume}
          className="action-button green focusable flex-grow text-lg"
        >
          <span
            className="icon-[iconamoon--player-play-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Resume
        </button>
        <button
          onClick={closeDialog}
          className="action-button green focusable flex-grow text-lg"
        >
          <span
            className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Close
        </button>
      </div>
    </dialog>
  );
}

export function BreakDialog({
  dialogRef,
  onContinue,
  onDone,
  planid,
  canContinue,
  length = "30 second",
}: {
  dialogRef: Ref<HTMLDialogElement>;
  onContinue: () => void;
  onDone: () => void;
  planid?: string;
  canContinue: boolean;
  length?: string;
}) {
  const closeDialog = useCallback(() => {
    if (dialogRef.current) {
      globalThis.handleCloseModal();
      if (dialogRef.current) {
        dialogRef.current.classList.add("close");
        setTimeout(() => {
          if (dialogRef.current) {
            dialogRef.current.classList.remove("close");
            dialogRef.current.close();
          }
        }, 150);
      }
    }
  }, [dialogRef]);

  const handleContinue = useCallback(() => {
    onContinue();
    closeDialog();
  }, [onContinue, closeDialog]);

  const handleDone = useCallback(() => {
    onDone();
    closeDialog();
  }, [onDone, closeDialog]);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby="break-title"
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
    >
      <header className="mt-2 text-left">
        <h3
          id="break-title"
          className="inline-flex items-center text-2xl font-semibold leading-6 text-neutral-900"
        >
          <span className="icon-[iconamoon--player-pause-fill] mr-1 size-8 text-violet-800" />
          {length} Pause!
        </h3>
      </header>
      <div className="mt-2 flex w-full flex-col gap-2 text-left text-neutral-700 sm:w-auto">
        <p>Take a short pause to reset, then continue when you’re ready.</p>
        {!!planid && (
          <>
            <p>This is a great time to practice your interleave spots!</p>
            <StartPracticingInterleave planid={planid} />
          </>
        )}
      </div>
      {canContinue ? (
        <div className="flex w-full flex-col flex-wrap gap-2 sm:flex-row-reverse">
          <button
            onClick={handleContinue}
            className="action-button green focusable flex-grow text-lg"
            disabled={!canContinue}
            type="button"
          >
            <span
              className="icon-[iconamoon--player-play-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Continue
          </button>
          <button
            onClick={handleDone}
            type="button"
            className="action-button amber focusable flex-grow text-lg"
          >
            <span
              className="icon-[iconamoon--player-stop-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Finish
          </button>
          {!!planid && <BackToPlan planid={planid} grow />}
        </div>
      ) : (
        <div className="mt-1 flex w-full justify-center py-2 text-lg font-medium">
          Enjoy your break!
        </div>
      )}
    </dialog>
  );
}
