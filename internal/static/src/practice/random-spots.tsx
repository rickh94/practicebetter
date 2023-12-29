import {
  StateUpdater,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "preact/hooks";
import { PracticeSummaryItem, RandomMode } from "../common";
import { BasicSpot } from "../validators";
import { ScaleCrossFadeContent } from "../ui/transitions";
import { CreateSpots } from "./create-spots";
import {
  BasicButton,
  BigAngryButton,
  BigHappyButton,
  BigSkyButton,
  GiantBasicButton,
  WarningButton,
} from "../ui/buttons";
import {
  HandRaisedIcon,
  HandThumbDownIcon,
  HandThumbUpIcon,
} from "@heroicons/react/24/solid";
import Summary from "./summary";
import { PracticeSpotDisplay } from "./practice-spot-display";
import dayjs from "dayjs";

export function RandomSpots({
  initialspots,
  pieceid,
  csrf,
}: {
  initialspots?: string;
  pieceid?: string;
  csrf?: string;
}) {
  const [spots, setSpots] = useState<BasicSpot[]>([]);
  const [skipSpotIds, setSkipSpotIds] = useState<string[]>([]);
  const [summary, setSummary] = useState<PracticeSummaryItem[]>([]);
  const [mode, setMode] = useState<RandomMode>("setup");
  const [startTime, setStartTime] = useState<Date | null>(null);

  const initialSpotIds = useMemo(
    function () {
      if (!initialspots) return [];
      const initialSpots: BasicSpot[] = JSON.parse(initialspots);
      if (!(initialSpots instanceof Array) || initialSpots.length === 0) {
        return [];
      }
      return initialSpots.map((spot) => spot.id);
    },
    [initialspots],
  );

  const finish = useCallback(
    function (finalSummary: PracticeSummaryItem[]) {
      setSummary(finalSummary);
      setMode("summary");
      document.removeEventListener(
        "UpdateSpotRemindersField",
        updateSpotRemindersField,
      );
    },
    [setMode, setSummary, startTime, pieceid, initialspots, csrf],
  );

  const updateSpotRemindersField = useCallback(
    function (event: CustomEvent) {
      const { id, text } = event.detail;
      setSpots((spots) =>
        spots.map((spot) =>
          spot.id === id ? { ...spot, textPrompt: text } : spot,
        ),
      );
    },
    [setSpots],
  );

  const startPracticing = useCallback(
    function () {
      setStartTime(new Date());
      setMode("practice");
      document.addEventListener(
        "UpdateSpotRemindersField",
        updateSpotRemindersField,
      );
    },
    [setMode, setStartTime, updateSpotRemindersField],
  );

  const backToSetup = useCallback(
    function () {
      setMode("setup");
      document.removeEventListener(
        "UpdateSpotRemindersField",
        updateSpotRemindersField,
      );
    },
    [setMode],
  );

  useEffect(
    function () {
      if (initialspots) {
        const initSpots: BasicSpot[] = JSON.parse(initialspots);
        setSpots(initSpots);
      }
    },
    [initialspots],
  );

  return (
    <div className="w-full">
      <ScaleCrossFadeContent
        component={
          {
            setup: (
              <SingleSetupForm
                setSpots={setSpots}
                spots={spots}
                submit={startPracticing}
              />
            ),
            practice: (
              <SinglePractice
                spots={spots}
                setup={backToSetup}
                finish={finish}
                skipSpotIds={skipSpotIds}
                setSkipSpotIds={setSkipSpotIds}
                pieceid={pieceid}
              />
            ),
            summary: (
              <Summary
                summary={summary}
                setup={() => setMode("setup")}
                practice={() => setMode("practice")}
                pieceHref={pieceid ? `/library/pieces/${pieceid}` : undefined}
                initialSpotIds={initialSpotIds}
                pieceid={pieceid}
                csrf={csrf}
                startTime={startTime}
              />
            ),
          }[mode]
        }
        id={mode}
      />
    </div>
  );
}

function SingleSetupForm({
  setSpots,
  spots,
  submit,
}: {
  setSpots: StateUpdater<BasicSpot[]>;
  submit: () => void;
  spots: BasicSpot[];
}) {
  const handleSubmit = useCallback(() => {
    if (spots.length === 0) {
      document.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            variant: "error",
            message: "Please enter at least one spot",
            title: "Missing Info",
            duration: 5000,
          },
        }),
      );
      return;
    }
    submit();
  }, [submit, spots]);

  return (
    <>
      <div className="flex w-full flex-col py-4">
        <div>
          <h1 className="py-1 text-left text-2xl font-bold">
            Single Random Spots
          </h1>
          <p className="text-left text-base">
            Enter your spots one at a time, or generate a bunch of spots at
            once.
          </p>
        </div>
        <div className="flex-shrink-0 flex-grow"></div>
      </div>
      <div className="flex w-full flex-col gap-y-4">
        <CreateSpots setSpots={setSpots} spots={spots} />
        <div className="col-span-full my-16 flex w-full items-center justify-center">
          <GiantBasicButton onClick={handleSubmit}>
            Start Practicing
          </GiantBasicButton>
        </div>
      </div>
    </>
  );
}

// TODO: add events to keep you from moving on while the reminders form is open
/*
 * Spot Promotion/Demotion rules
 * - just five excellents, recommend promotion beyond day three
 * - always evict after five net excellents (minus poor)
 * - after day five, demote if no excellents
 * - always evict after three poors
 * - after day three, demote after three poors
 */

function SinglePractice({
  spots,
  setup,
  finish,
  pieceid,
}: {
  spots: BasicSpot[];
  setup: () => void;
  finish: (summary: PracticeSummaryItem[]) => void;
  skipSpotIds: string[];
  setSkipSpotIds: StateUpdater<string[]>;
  pieceid?: string;
}) {
  const [currentSpotIdx, setCurrentSpotIdx] = useState(
    Math.floor(Math.random() * spots.length),
  );
  const [practiceSummary, setPracticeSummary] = useState<
    Map<string, { excellent: number; fine: number; poor: number }>
  >(new Map());
  // This counter ensures that the animation runs, even if the same spot is generated twice in a row.
  const [counter, setCounter] = useState(0);
  const [skipSpotIds, setSkipSpotIds] = useState<string[]>([]);
  const [lastTwoSpots, setLastTwoSpots] = useState<string[]>([]);
  const topRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (topRef.current) {
      window.scrollTo(0, topRef.current.offsetTop);
    }
  }, [topRef.current]);

  const addSpotRep = useCallback(
    function (id: string | undefined, quality: "excellent" | "fine" | "poor") {
      if (!id || !quality) {
        return;
      }
      let summary = practiceSummary.get(id) ?? {
        excellent: 0,
        fine: 0,
        poor: 0,
      };
      summary[quality] += 1;
      practiceSummary.set(id, summary);
      setPracticeSummary(practiceSummary);
      return summary;
    },
    [setPracticeSummary, practiceSummary],
  );

  // TODO: handle adding rep correctly when hitting done
  const handleDone = useCallback(
    function () {
      const finalSummary: PracticeSummaryItem[] = [];
      for (const spot of spots) {
        let results = practiceSummary.get(spot.id) ?? {
          excellent: 0,
          fine: 0,
          poor: 0,
        };
        let day = 0;
        if (spot.stageStarted) {
          let stageStarted = dayjs.unix(spot.stageStarted).tz(dayjs.tz.guess());
          let now = dayjs().tz(dayjs.tz.guess());
          day = now.diff(stageStarted, "day");
        }
        finalSummary.push({
          name: spot.name ?? "Missing spot name",
          reps: results.excellent + results.fine + results.poor,
          excellent: results.excellent,
          fine: results.fine,
          poor: results.poor,
          id: spot.id ?? "Missing spot id",
          day,
        });
      }
      finish(finalSummary);
    },
    [practiceSummary, finish, spots],
  );

  const nextSpot = useCallback(
    function (nextSkipSpotIds: string[]) {
      setCounter((curr) => curr + 1);
      // addSpotRep(spots[currentSpotIdx]?.id);
      if (nextSkipSpotIds.length >= spots.length) {
        document.dispatchEvent(
          new CustomEvent("ShowAlert", {
            detail: {
              message: "You practiced every spot!",
              title: "Practicing Complete",
              variant: "success",
              duration: 3000,
            },
          }),
        );
        handleDone();
        return;
      }
      let nextSpotIdx = Math.floor(Math.random() * spots.length);
      let nextSpotId = spots[nextSpotIdx]?.id;
      while (
        !nextSpotId ||
        (nextSpotId && nextSkipSpotIds.includes(nextSpotId)) ||
        // check if the last two spots are the same and also the same as the next spot
        // but only if the skip spots is more than one smaller than the spots, otherwise
        // there is only one spot left and it will infinitely loop
        (nextSpotId &&
          spots.length - 1 > nextSkipSpotIds.length &&
          lastTwoSpots[0] === lastTwoSpots[1] &&
          lastTwoSpots[1] === nextSpotId)
      ) {
        nextSpotIdx = Math.floor(Math.random() * spots.length);
        nextSpotId = spots[nextSpotIdx]?.id;
      }
      setCurrentSpotIdx(nextSpotIdx);
      setLastTwoSpots([nextSpotId, lastTwoSpots[0]]);
    },
    [addSpotRep, currentSpotIdx, handleDone, spots, lastTwoSpots],
  );

  const evictSpot = useCallback(
    function (spotId: string) {
      // going to need a copy of this because the it won't be updated by setstate until after the function finishes
      const newSkipSpotIds = [...skipSpotIds];
      if (spotId) {
        newSkipSpotIds.push(spotId);
      }
      setSkipSpotIds(newSkipSpotIds);
      return newSkipSpotIds;
    },
    [spots, skipSpotIds],
  );

  const handleExcellent = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      const summary = addSpotRep(currentSpotId, "excellent");
      let nextSkipSpotIds = skipSpotIds;
      if (summary.excellent - summary.poor > 4) {
        nextSkipSpotIds = evictSpot(currentSpotId);
      }
      nextSpot(nextSkipSpotIds);
    },
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, nextSpot, evictSpot],
  );

  const handleFine = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      addSpotRep(currentSpotId, "fine");
      nextSpot(skipSpotIds);
    },
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, nextSpot],
  );

  const handlePoor = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      const summary = addSpotRep(currentSpotId, "poor");
      let nextSkipSpotIds = skipSpotIds;
      if (summary.poor > 2) {
        nextSkipSpotIds = evictSpot(currentSpotId);
      }
      nextSpot(nextSkipSpotIds);
    },
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, nextSpot, evictSpot],
  );

  return (
    <div className="relative w-full" ref={topRef}>
      <div className="flex w-full flex-col items-center justify-center gap-2 pt-2">
        <div className="text-2xl font-semibold text-neutral-700">
          Practicing:
        </div>
        <div className="relative w-full py-4">
          <ScaleCrossFadeContent
            component={
              <PracticeSpotDisplay
                spot={spots[currentSpotIdx]}
                pieceid={pieceid}
              />
            }
            id={`${currentSpotIdx}-${counter}`}
          />
        </div>
        <div className="flex w-full flex-col justify-center gap-2 px-4 pt-12 sm:flex-row-reverse sm:px-0">
          <BigHappyButton
            type="button"
            onClick={handleExcellent}
            className="gap-2"
          >
            <HandThumbUpIcon className="-ml-1 size-6" />
            Excellent
          </BigHappyButton>
          <BigSkyButton type="button" onClick={handleFine} className="gap-2">
            <HandRaisedIcon className="-mr-1 size-6" />
            Fine
          </BigSkyButton>
          <BigAngryButton type="button" onClick={handlePoor} className="gap-2">
            <HandThumbDownIcon className="-mr-1 size-6" />
            Poor
          </BigAngryButton>
        </div>
        <div className="flex justify-center gap-4 pt-8">
          <BasicButton onClick={setup}>← Back to setup</BasicButton>
          <WarningButton onClick={handleDone}>Done</WarningButton>
        </div>
      </div>
    </div>
  );
}
