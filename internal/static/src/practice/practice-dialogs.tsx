import { Ref, useCallback, useState } from "preact/hooks";
import { HappyButton, AngryButton, WarningButton } from "../ui/buttons";
import { InterleaveSpotsList } from "../ui/plan-components";
import { BackToPlan } from "../ui/links";

export function ResumeDialog({
  dialogRef,
  onResume,
}: {
  dialogRef: Ref<HTMLDialogElement>;
  onResume: () => void;
}) {
  const closeDialog = useCallback(
    function () {
      if (dialogRef.current) {
        globalThis.handleCloseModal();
        dialogRef.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (dialogRef.current) {
              dialogRef.current.classList.remove("close");
              dialogRef.current.close();
            }
          });
        });
      }
    },
    [dialogRef.current],
  );

  const handleResume = useCallback(
    function () {
      onResume();
      closeDialog();
    },
    [onResume, closeDialog],
  );

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
        <HappyButton grow onClick={handleResume} className="text-lg">
          <span
            className="icon-[iconamoon--player-play-thin] -ml-1 size-5"
            aria-hidden="true"
          ></span>
          Resume
        </HappyButton>
        <AngryButton grow onClick={closeDialog} className="text-lg">
          <span
            className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
            aria-hidden="true"
          ></span>
          Close
        </AngryButton>
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
  const [shouldFetch, setShouldFetch] = useState(true);
  const closeDialog = useCallback(
    function () {
      if (dialogRef.current) {
        globalThis.handleCloseModal();
        dialogRef.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (dialogRef.current) {
              dialogRef.current.classList.remove("close");
              dialogRef.current.close();
            }
          });
        });
        setShouldFetch(true);
      }
    },
    [dialogRef.current],
  );

  const handleContinue = useCallback(
    function () {
      onContinue();
      closeDialog();
    },
    [onContinue, closeDialog],
  );

  const handleDone = useCallback(
    function () {
      onDone();
      closeDialog();
    },
    [onDone, closeDialog],
  );

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby="break-title"
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
    >
      <header className="mt-2 text-center sm:text-left">
        <h3
          id="break-title"
          className="text-2xl font-semibold leading-6 text-neutral-900"
        >
          Time for a {length} Break
        </h3>
      </header>
      <div className="mt-2 flex w-full flex-col gap-2 text-left text-neutral-700 sm:w-auto">
        <p>
          You have been practicing for a while. Give your body and mind a short
          break, then continue.
        </p>
        {!!planid && (
          <>
            <p>This is a great time to practice your interleave spots!</p>
            <InterleaveSpotsList
              planid={planid}
              shouldFetch={shouldFetch}
              setShouldFetch={setShouldFetch}
            />
            <p>
              You can also go back to your practice plan and resume this later.
            </p>
          </>
        )}
      </div>
      {canContinue ? (
        <div className="flex w-full flex-col flex-wrap gap-2 sm:flex-row-reverse">
          <HappyButton
            grow
            onClick={handleContinue}
            className="text-lg"
            disabled={!canContinue}
          >
            <span
              className="icon-[iconamoon--player-play-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>
            Continue
          </HappyButton>
          <WarningButton grow onClick={handleDone}>
            <span
              className="icon-[iconamoon--player-stop-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>
            Finish
          </WarningButton>
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
