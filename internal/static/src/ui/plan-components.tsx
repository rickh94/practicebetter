import { BackToPlan, Link } from "./links";
import {
  type StateUpdater,
  useCallback,
  useEffect,
  useRef,
} from "preact/hooks";
import * as htmx from "htmx.org";
import { forwardRef } from "preact/compat";
import { type Ref } from "preact";

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
          <InterleaveSpotsList planid={planid} shouldFetch={true} />
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

export const InterleaveSpotsList = forwardRef(
  (
    props: {
      planid?: string;
      shouldFetch?: boolean;
      setShouldFetch?: StateUpdater<boolean>;
    },
    ref: Ref<HTMLDetailsElement>,
  ) => {
    const interleaveSpotsRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
      if (interleaveSpotsRef.current && props.shouldFetch) {
        htmx
          .ajax(
            "GET",
            `/library/plans/${props.planid}/interleave`,
            interleaveSpotsRef.current,
          )
          .then(() => props.setShouldFetch?.(false))
          .catch((err) => console.error(err));
      }
    }, [props.shouldFetch, props.setShouldFetch, props]);

    return (
      <details className="my-1 w-full" ref={ref}>
        <summary className="focusable indigo flex cursor-pointer select-none items-center justify-between gap-1 rounded-xl border border-indigo-400 bg-indigo-200 py-2 pl-4 pr-2 font-medium text-indigo-800 shadow-sm shadow-purple-900/30 transition duration-200 hover:border-indigo-500 hover:bg-indigo-300 hover:shadow-indigo-900/50">
          <div className="flex items-center gap-2 focus:outline-none">
            <span className="icon-[iconamoon--bookmark-thin] -ml-1 size-5" />
            Interleave Spots
          </div>
          <span
            className="summary-icon icon-[iconamoon--arrow-right-6-circle-thin] size-6 transition-transform"
            aria-hidden="true"
          />
        </summary>
        {props.planid ? (
          <div ref={interleaveSpotsRef} className="w-full py-2">
            Loading Interleave Spots...
          </div>
        ) : (
          <div className="w-full py-2">No interleave spots</div>
        )}
      </details>
    );
  },
);
