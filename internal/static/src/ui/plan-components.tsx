import { BackToPlan, Link } from "./links";
import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import * as htmx from "htmx.org/dist/htmx";
import dayjs from "dayjs";
import duration from "dayjs/plugin/duration";

// on mount check for break async, then set which dialog to show in state and show that on click,
// once they take a break, show a resume from break dialog that has a go on button
//
dayjs.extend(duration);

export function NextPlanItem({
  planid,
  csrf,
}: {
  planid: string;
  csrf?: string;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const openDialog = useCallback(() => {
    dialogRef.current?.showModal();
    globalThis.handleShowModal();
  }, [dialogRef]);

  const [needsBreak, setNeedsBreak] = useState(false);
  const [showBreak, setShowBreak] = useState(false);

  useEffect(() => {
    fetch("/library/break")
      .then((res) => {
        if (res.ok) {
          res
            .text()
            .then((text) => setNeedsBreak(text === "true"))
            .catch(console.error);
        } else {
          console.error(res.statusText);
        }
      })
      .catch(console.error);
  }, [dialogRef]);

  const handleGoOnClick = useCallback(() => {
    if (needsBreak) {
      setShowBreak(true);
    } else {
      openDialog();
    }
  }, [needsBreak, openDialog]);

  return (
    <>
      <button
        className="focusable action-button green"
        onClick={handleGoOnClick}
      >
        Go On
        <span
          className="icon-[iconamoon--player-next-thin] -mr-1 size-5"
          aria-hidden="true"
        />
      </button>
      {csrf ? (
        <BreakDialog
          open={showBreak}
          doNext={() => {
            setTimeout(() => openDialog(), 100);
          }}
          csrf={csrf}
        />
      ) : null}
      <dialog
        ref={dialogRef}
        id="practice-next-dialog"
        aria-labelledby="practice-next-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] p-4 text-left sm:max-w-xl"
      >
        <header className="flex h-8 flex-shrink-0 text-left">
          <h3
            id="practice-next-title"
            className="inline-block text-2xl font-semibold leading-6 text-neutral-900"
          >
            Next Practice Item
          </h3>
        </header>
        <div className="flex w-full flex-shrink-0 flex-col gap-2 text-left text-neutral-700 sm:w-[32rem]">
          <p className="text-sm">
            You can take this opportunity to go through your interleave spots,
            or move on.
          </p>
          <StartPracticingInterleave planid={planid} />
        </div>
        <div className="grid w-full grid-cols-1 gap-2 xs:grid-cols-2">
          <BackToPlan grow planid={planid} />
          <Link
            href={`/library/plans/${planid}/next`}
            className="action-button green focusable flex-grow"
          >
            Go On
            <span
              className="icon-[iconamoon--player-next-thin] -mr-1 size-5"
              aria-hidden="true"
            />
          </Link>
        </div>
      </dialog>
    </>
  );
}

export const StartPracticingInterleave = (props: { planid?: string }) => {
  const dialogRef = useRef<HTMLDialogElement>(null);

  const closeModal = useCallback(() => {
    dialogRef.current?.close();
    globalThis.handleCloseModal();
  }, []);

  const startPracticing = useCallback(() => {
    htmx
      .ajax(
        "GET",
        `/library/plans/${props.planid}/interleave/start?goOn=true`,
        {
          target: "#interleave-spot-dialog-contents",
          swap: "innerHTML transition:true",
        },
      )
      .then(() => {
        if (dialogRef.current) {
          dialogRef.current.showModal();
          globalThis.handleShowModal();
          globalThis.addEventListener("FinishedInterleave", closeModal);
        }
      })
      .catch(() => {
        console.error("Could not get interleave spots");
        globalThis.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              message: "Could not get interleave spots",
              title: "Error",
              variant: "error",
              duration: 3000,
            },
          }),
        );
      });
    return () =>
      globalThis.removeEventListener("FinishedInterleave", closeModal);
  }, [closeModal, props.planid]);

  return (
    <>
      <button
        onClick={startPracticing}
        class="action-button indigo focusable px-4 text-lg"
      >
        Practice Interleave
        <span
          class="icon-[iconamoon--player-play-thin] -ml-1 size-5"
          aria-hidden="true"
        />
      </button>
      <dialog
        ref={dialogRef}
        id="interleave-spot-dialog"
        class="clear flex flex-col gap-2 bg-transparent p-4 text-left focus:outline-none"
      >
        <div
          id="interleave-spot-dialog-contents"
          class="w-huge overflow-x-clip p-0"
        >
          <span class="rounded-xl bg-white p-4">
            Loading Interleave Spot...
          </span>
        </div>
        <div class="mx-4 mt-4 overflow-x-clip rounded-xl p-4 sm:mx-auto sm:w-96">
          <button
            class="amber action-button focusable mx-auto"
            onClick={closeModal}
          >
            <span
              class="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Close
          </button>
        </div>
      </dialog>
    </>
  );
};

export function BreakDialog({
  open = false,
  csrf,
  doNext,
}: {
  open?: boolean;
  csrf: string;
  doNext?: () => void;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [breakState, setBreakState] = useState<
    "prepare" | "break" | "finished"
  >("prepare");
  const [secondsRemaining, setSecondsRemaining] = useState(0);
  const [breakEndTime, setBreakEndTime] = useState(0);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const openDialog = useCallback(() => {
    dialogRef.current?.showModal();
    globalThis.handleShowModal();
  }, [dialogRef]);

  const closeDialog = useCallback(() => {
    dialogRef.current?.close();
    globalThis.handleCloseModal();
  }, []);

  useEffect(() => {
    if (open) {
      openDialog();
    } else {
      closeDialog();
    }
  }, [closeDialog, open, openDialog]);

  const startBreak = useCallback(async () => {
    setBreakState("break");
    setSecondsRemaining(3 * 60);
    setBreakEndTime(Date.now() + 3 * 60 * 1000);
    if (countdownRef.current) {
      clearInterval(countdownRef.current);
    }
    countdownRef.current = setInterval(() => {
      setSecondsRemaining((secondsRemaining) => secondsRemaining - 1);
    }, 1000);
    // save that they took a break in case they refresh or something weird happens.
    try {
      const res = await fetch("/library/break", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
      });
      if (!res.ok) {
        console.error(res);
        alert("Could not record break");
      }
    } catch (err) {
      alert("Could not record break");
      console.error(err);
    }
  }, [csrf]);

  const finishBreak = useCallback(() => {
    if (countdownRef.current) {
      clearInterval(countdownRef.current);
    }
    setBreakState("finished");
    setSecondsRemaining(0);
    setBreakEndTime(0);
  }, []);

  const continuePracticing = useCallback(() => {
    closeDialog();
    doNext?.();
    if (breakState === "finished") {
      // update the break time to when they returned so long breaks don't get weird
      fetch("/library/break", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
      }).catch(console.error);
    }
  }, [breakState, closeDialog, csrf, doNext]);

  useEffect(() => {
    if (breakState !== "break") {
      return;
    }
    if (secondsRemaining <= 0 || breakEndTime - Date.now() <= 0) {
      finishBreak();
    }
  }, [secondsRemaining, closeDialog, finishBreak, breakState, breakEndTime]);

  return (
    <dialog
      ref={dialogRef}
      id="take-break-dialog"
      aria-labelledby="take-break-title"
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] p-4 text-left sm:max-w-xl"
    >
      <header className="flex h-8 flex-shrink-0 text-left">
        <h3
          id="take-break-title"
          className="inline-block text-2xl font-semibold leading-6 text-neutral-900"
        >
          {breakTitle(breakState)}
        </h3>
      </header>
      <div className="flex w-full flex-shrink-0 flex-col gap-2 py-2 text-left text-neutral-700 sm:w-[32rem]">
        {breakText(breakState)}
      </div>
      <div className="flex w-full flex-col-reverse gap-2 sm:grid sm:grid-cols-2">
        <BreakActions
          breakState={breakState}
          secondsRemaining={secondsRemaining}
          startBreak={startBreak}
          continuePracticing={continuePracticing}
        />
      </div>
    </dialog>
  );
}

function breakTitle(breakState: "prepare" | "break" | "finished") {
  switch (breakState) {
    case "prepare":
      return "Need a Break?";
    case "break":
      return "Taking a Break...";
    case "finished":
      return "Continue Practicing";
  }
}

function breakText(breakState: "prepare" | "break" | "finished") {
  switch (breakState) {
    case "prepare":
      return "You’ve been practicing for a while. Take a break to keep yourself fresh";
    case "break":
      return "Come back in a few minutes to continue practicing";
    case "finished":
      return "Your break is over, time to keep going.";
  }
}

function BreakActions({
  breakState,
  secondsRemaining,
  startBreak,
  continuePracticing,
}: {
  breakState: "prepare" | "break" | "finished";
  secondsRemaining: number;
  startBreak: () => void;
  continuePracticing: () => void;
}) {
  switch (breakState) {
    case "prepare":
      return (
        <>
          <button
            className="action-button focusable red"
            onClick={continuePracticing}
          >
            <span
              className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Dismiss
          </button>

          <button
            className="action-button focusable green"
            onClick={startBreak}
          >
            <span
              className="icon-[iconamoon--clock-duotone] -ml-1 size-5"
              aria-hidden="true"
            />
            Take a Break
          </button>
        </>
      );
    case "break":
      return (
        <>
          <div className="col-span-full text-center text-lg font-medium">
            Come back in{" "}
            <span className="font-bold">
              {dayjs.duration(secondsRemaining, "s").format("m:ss")}
            </span>
          </div>
        </>
      );
    case "finished":
      return (
        <>
          <button
            className="action-button focusable green col-span-full"
            onClick={continuePracticing}
          >
            <span
              className="icon-[iconamoon--player-play-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Continue
          </button>
        </>
      );
  }
}
