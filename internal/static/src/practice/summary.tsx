import { cn, type PracticeSummaryItem } from "../common";
import { BackToPiece, BackToPlan } from "../ui/links";
import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { useAutoAnimate } from "@formkit/auto-animate/preact";
import { NextPlanItem } from "../ui/plan-components";

// TODO: improve table layout and appearance
// TODO: prevent escape closing window or use it to reject recommendations, also maybe add reject button
// TODO: implement some kind of sorting, or find a way to retain entered order,
export default function Summary({
  summary,
  setup,
  practice,
  pieceid,
  csrf,
  startTime,
  initialSpotIds,
  planid = "",
}: {
  summary: PracticeSummaryItem[];
  setup: () => void;
  practice: () => void;
  pieceid?: string;
  csrf?: string;
  startTime?: Date;
  initialSpotIds?: string[];
  planid?: string;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [promotionSpots, setPromotionSpots] = useState<PracticeSummaryItem[]>(
    [],
  );
  const [demotionSpots, setDemotionSpots] = useState<PracticeSummaryItem[]>([]);
  const [hasSetup, setHasSetup] = useState(false);

  const submit = useCallback(() => {
    if (pieceid && csrf && startTime) {
      const seenIds = new Set();
      const spots: { id: string; promote: boolean; demote: boolean }[] = [];
      for (const spot of promotionSpots) {
        if (seenIds.has(spot.id)) continue;
        if (!initialSpotIds?.includes(spot.id)) continue;
        spots.push({ id: spot.id, promote: true, demote: false });
        seenIds.add(spot.id);
      }
      for (const spot of demotionSpots) {
        if (seenIds.has(spot.id)) continue;
        if (!initialSpotIds?.includes(spot.id)) continue;
        spots.push({ id: spot.id, promote: false, demote: true });
        seenIds.add(spot.id);
      }
      for (const spot of summary) {
        if (seenIds.has(spot.id)) continue;
        if (!initialSpotIds?.includes(spot.id)) continue;
        spots.push({ id: spot.id, promote: false, demote: false });
        seenIds.add(spot.id);
      }
      const durationMinutes = Math.ceil(
        (new Date().getTime() - startTime.getTime()) / 1000 / 60,
      );
      globalThis.dispatchEvent(
        new CustomEvent("FinishedSpotPracticing", {
          detail: {
            spots,
            durationMinutes,
            csrf,
            endpoint: `/library/pieces/${pieceid}/practice/random-single`,
          },
        }),
      );
    }
  }, [
    promotionSpots,
    demotionSpots,
    pieceid,
    csrf,
    summary,
    initialSpotIds,
    startTime,
  ]);

  const close = useCallback(() => {
    globalThis.handleCloseModal();
    if (dialogRef.current) {
      if (dialogRef.current) {
        dialogRef.current.classList.add("close");
        setTimeout(() => {
          dialogRef.current?.close();
          dialogRef.current?.classList.remove("close");
        }, 150);
      }
    }
  }, []);

  const savePromotions = useCallback(() => {
    submit();
    close();
  }, [close, submit]);

  const rejectPromotions = useCallback(() => {
    setPromotionSpots([]);
    setDemotionSpots([]);
    submit();
    close();
  }, [close, submit, setPromotionSpots, setDemotionSpots]);

  const removePromotionSpot = useCallback(
    (id: string) => {
      setPromotionSpots((prev) => prev.filter((spot) => spot.id !== id));
    },
    [setPromotionSpots],
  );

  const removeDemotionSpot = useCallback(
    (id: string) => {
      setDemotionSpots((prev) => prev.filter((spot) => spot.id !== id));
    },
    [setDemotionSpots],
  );

  /*
   * Spot Promotion/Demotion rules
   * - just three excellents, recommend promotion beyond day four
   * - always evict after five net excellents (minus poor)
   * - after day five, demote if no excellents
   * - always evict after three poors
   * - after day three, demote after three poors
   */
  useEffect(() => {
    if (hasSetup) {
      return;
    }
    console.log(summary);
    setHasSetup(true);
    const promote: PracticeSummaryItem[] = [];
    const demote: PracticeSummaryItem[] = [];
    for (const item of summary) {
      if (
        item.excellent > 2 &&
        item.poor === 0 &&
        item.fine < 2 &&
        item.day > 5
      ) {
        promote.push(item);
      } else if (
        item.poor > 2 ||
        (item.excellent === 0 && item.fine > 0 && item.poor > 0 && item.day > 6)
      ) {
        demote.push(item);
      }
    }
    setPromotionSpots(promote);
    setDemotionSpots(demote);
    if (dialogRef.current && (promote.length > 0 || demote.length > 0)) {
      globalThis.handleShowModal();
      dialogRef.current.showModal();
    } else {
      submit();
    }
  }, [
    dialogRef,
    setPromotionSpots,
    setDemotionSpots,
    summary,
    submit,
    hasSetup,
  ]);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  const [promotionListParent] = useAutoAnimate();
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  const [demotionListParent] = useAutoAnimate();

  // TODO: change save button conditionally
  return (
    <>
      <dialog
        ref={dialogRef}
        aria-labelledby="promote-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="promote-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Recommendations
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          Based on your practicing, here are some recommendations for your
          spots. Click the{" "}
          <span className="-mb-2 inline-flex w-4 items-center justify-center">
            <span
              className="icon-[iconamoon--sign-minus-circle-thin] size-4 text-red-500"
              aria-hidden="true"
            />
            <span className="sr-only">Remove Button</span>
          </span>{" "}
          to prevent the change and keep the spot in Random practicing.
        </div>
        <div className="flex flex-col-reverse gap-2  sm:grid sm:grid-cols-2">
          <div className="flex flex-col gap-2 rounded-xl bg-amber-500/10 p-2">
            <h4 className="text-center text-lg font-bold">Demote Spots</h4>
            <p className="text-sm text-neutral-800">
              These spots could use a little more attention. Let’s send them
              back to repeat practicing for now.
            </p>
            {demotionSpots.length > 0 ? (
              <ul
                className="flex list-none flex-col gap-2"
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
                ref={demotionListParent}
              >
                {demotionSpots.map((item) => (
                  <li
                    key={item.id}
                    className="flex items-center justify-between"
                  >
                    {item.name}
                    <button
                      onClick={() => removeDemotionSpot(item.id)}
                      className="focusable flex items-center justify-center rounded-full p-1 text-red-500 hover:bg-red-500/10 hover:text-red-700"
                    >
                      <span
                        className="icon-[iconamoon--sign-minus-circle-thin] size-5"
                        aria-hidden="true"
                      />
                      <span className="sr-only">Remove</span>
                    </button>
                  </li>
                ))}
              </ul>
            ) : (
              <p>No spots to demote today</p>
            )}
          </div>
          <div className="flex flex-col gap-2 rounded-xl bg-sky-500/10 p-2">
            <h4 className="text-center text-lg font-bold">Promote Spots</h4>
            <p className="text-sm text-neutral-800">
              These spots are going really well, we can move them on to
              interleaved practicing!{" "}
            </p>
            {promotionSpots.length > 0 ? (
              <ul
                className="flex list-none flex-col gap-2"
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
                ref={promotionListParent}
              >
                {promotionSpots.map((item) => (
                  <li
                    key={item.id}
                    className="flex items-center justify-between"
                  >
                    {item.name}
                    <button
                      onClick={() => removePromotionSpot(item.id)}
                      className="focusable flex items-center justify-center rounded-full p-1 text-red-500 hover:bg-red-500/10 hover:text-red-700"
                    >
                      <span
                        className="icon-[iconamoon--sign-minus-circle-thin] size-5"
                        aria-hidden="true"
                      />
                      <span className="sr-only">Remove</span>
                    </button>
                  </li>
                ))}
              </ul>
            ) : (
              <p>No spots to promote today</p>
            )}
          </div>
        </div>
        <div className="mt-2 flex w-full flex-col-reverse gap-2 xs:grid xs:grid-cols-2">
          <button
            onClick={rejectPromotions}
            className="action-button focusable red w-full flex-grow text-lg"
          >
            <span
              className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Do Nothing
          </button>
          <button
            onClick={savePromotions}
            className="action-button green focusable w-full flex-grow text-lg"
          >
            <span
              className="icon-[iconamoon--check-circle-1-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Accept
          </button>
        </div>
      </dialog>
      <div className="flex w-full flex-col justify-center gap-x-8 gap-y-2 px-4 pt-4 sm:flex-row sm:gap-x-6 sm:px-0">
        <SummaryActions
          pieceid={pieceid}
          planid={planid}
          setup={setup}
          practice={practice}
          csrf={csrf}
        />
      </div>
      <h2 className="w-full pt-12 text-center text-2xl font-semibold">
        Practice Summary
      </h2>
      <div className="flex w-full flex-col items-center justify-center gap-2 pt-4">
        <table className="hidden w-full divide-y divide-neutral-700 sm:table">
          <thead className="w-full">
            <tr>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:pl-0"
              >
                Spot Name
              </th>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-green-800 sm:pl-0"
              >
                Excellent
              </th>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-sky-800 sm:pl-0"
              >
                Fine
              </th>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-red-800 sm:pl-0"
              >
                Poor
              </th>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:pl-0"
              >
                Total
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-neutral-700">
            {summary.map(({ name, reps, id, excellent, fine, poor }, idx) => (
              <tr
                key={id}
                className={`${idx % 2 === 0 && "bg-neutral-700/10"}`}
              >
                <td className="whitespace-nowrap py-2 pl-4 pr-3 text-center font-medium text-neutral-900 sm:pl-0">
                  {name}
                </td>
                <td className="whitespace-nowrap px-3 py-2 text-center text-green-800">
                  {excellent}
                </td>
                <td className="whitespace-nowrap px-3 py-2 text-center text-sky-800">
                  {fine}
                </td>
                <td className="whitespace-nowrap px-3 py-2 text-center text-red-800">
                  {poor}
                </td>
                <td className="whitespace-nowrap px-3 py-2 text-center text-neutral-800">
                  {reps}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <ul className="flex w-full list-none flex-col divide-y divide-neutral-700 border-y border-y-neutral-700 sm:hidden">
          {summary.map(({ name, reps, id, excellent, fine, poor }, idx) => (
            <li
              className={cn(
                "flex flex-col gap-1 px-3 py-2",
                idx % 2 === 0 && "bg-neutral-700/10",
              )}
              key={id}
            >
              <span className="flex gap-1">
                Spot:
                <strong className="font-bold">{name}</strong>
              </span>
              <span className="flex w-full flex-wrap justify-between gap-4">
                <span className="flex max-w-xs flex-grow flex-wrap justify-between gap-2">
                  <span className="flex gap-1 text-green-800">
                    Excellent:{" "}
                    <strong className="font-bold">{excellent}</strong>
                  </span>
                  <span className="flex gap-1 text-sky-800">
                    Fine: <strong className="font-bold">{fine}</strong>
                  </span>
                  <span className="flex gap-1 text-red-800">
                    Poor: <strong className="font-bold">{poor}</strong>
                  </span>
                </span>
                <span className="flex gap-1 self-end justify-self-end text-black">
                  Total: <strong className="font-bold">{reps}</strong>
                </span>
              </span>
            </li>
          ))}
        </ul>
      </div>
    </>
  );
}

export function SummaryActions({
  planid,
  setup,
  practice,
  pieceid,
  csrf,
}: {
  planid?: string;
  pieceid?: string;
  setup: () => void;
  practice: () => void;
  csrf?: string;
}) {
  if (planid) {
    return (
      <>
        {pieceid && <BackToPiece pieceid={pieceid} />}
        {planid && <BackToPlan planid={planid} />}
        <NextPlanItem planid={planid} csrf={csrf} />
      </>
    );
  }
  return (
    <>
      <button
        onClick={setup}
        type="button"
        className="action-button red focusable"
      >
        <span
          className="icon-[iconamoon--settings-thin] -ml-1 size-5"
          aria-hidden="true"
        />
        Back to Setup
      </button>
      <button
        onClick={practice}
        type="button"
        className="action-button violet focusable"
      >
        <span
          className="icon-[iconamoon--music-2-thin] -ml-1 size-5"
          aria-hidden="true"
        />
        Practice More
      </button>
      {pieceid && <BackToPiece pieceid={pieceid} />}
    </>
  );
}
