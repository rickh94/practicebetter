import { BackToPlan, Link } from "./links";
import { useCallback, useRef } from "preact/hooks";
import * as htmx from "htmx.org";

export function NextPlanItem({ planid }: { planid: string }) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const openDialog = useCallback(() => {
    dialogRef.current?.showModal();
    globalThis.handleShowModal();
  }, [dialogRef]);
  return (
    <>
      <button className="focusable action-button green" onClick={openDialog}>
        Go On
        <span
          className="icon-[iconamoon--player-next-thin] -mr-1 size-5"
          aria-hidden="true"
        />
      </button>
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
            className="focusable action-button green flex-grow"
            href={`/library/plans/${planid}/next`}
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

// TODO: manage the thing so that on the last one it goes on or just doesn't close both modals

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
        <div class="mx-4 mt-4 w-full overflow-x-clip rounded-xl border border-neutral-500 bg-white p-0 sm:mx-auto sm:w-96">
          <button
            class="amber action-button focusable w-full"
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
