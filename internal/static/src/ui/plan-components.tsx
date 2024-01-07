import { BackToPlan, Link } from "./links";
import { useCallback, useEffect, useRef } from "preact/hooks";
import * as htmx from "htmx.org";

export function NextPlanItem({ planid }: { planid?: string }) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const openDialog = useCallback(() => {
    dialogRef.current?.showModal();
  }, [dialogRef.current]);
  return (
    <>
      <button
        className="focusable action-button bg-green-700/10 text-green-800 hover:bg-green-700/20"
        onClick={openDialog}
      >
        Go On
        <span
          className="icon-[heroicons--arrow-right-circle-solid] -mr-1 size-5"
          aria-hidden="true"
        />
      </button>
      <dialog
        ref={dialogRef}
        aria-labelledby="practice-next-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-2 text-left sm:max-w-xl"
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
          <p className="inline-block">What to do next?</p>
          <ul className="block list-disc pl-4">
            <li>Practice your interleave spots</li>
            <li>Take a short break</li>
            <li>Go on to the next item</li>
            <li>Go back to your practice plan</li>
          </ul>
          <InterleaveSpotsList planid={planid} />
        </div>
        <div className="grid w-full grid-cols-1 gap-2 xs:grid-cols-2">
          <BackToPlan grow planid={planid} />
          <Link
            className="focusable action-button flex-grow bg-green-700/10 text-green-800 hover:bg-green-700/20"
            href={`/library/plans/${planid}/next`}
          >
            Go On
            <span
              className="icon-[heroicons--arrow-right-circle-solid] -mr-1 size-5"
              aria-hidden="true"
            />
          </Link>
        </div>
      </dialog>
    </>
  );
}

export function InterleaveSpotsList({ planid }: { planid?: string }) {
  const interleaveSpotsRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (interleaveSpotsRef.current) {
      htmx.process(interleaveSpotsRef.current);
    }
  }, [interleaveSpotsRef.current]);

  return (
    <details className="my-1 w-full">
      <summary className="focusable flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-indigo-500/50 py-2 pl-4 pr-2 font-semibold text-indigo-800 transition duration-200 hover:bg-indigo-300/50 focus:outline-none">
        <div className="flex items-center gap-2 focus:outline-none">
          <span className="icon-[heroicons--bookmark-solid] -ml-1 size-5" />
          Interleave Spots
        </div>
        <span
          className="summary-icon icon-[heroicons--chevron-right] size-6 transition-transform"
          aria-hidden="true"
        />
      </summary>
      {!!planid ? (
        <div
          ref={interleaveSpotsRef}
          hx-trigger="revealed"
          hx-swap="innterHTML transition:true"
          hx-get={`/library/plans/${planid}/interleave`}
          className="w-full py-2"
        >
          Loading Interleave Spots...
        </div>
      ) : (
        <div className="w-full py-2">No interleave spots</div>
      )}
    </details>
  );
}
