import {
  StateUpdater,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "preact/hooks";
import { PracticeSummaryItem, RandomMode, cn } from "../common";
import { BasicSpot } from "../validators";
import { ScaleCrossFadeContent } from "../ui/transitions";
import { CreateSpots } from "./create-spots";
import {
  BasicButton,
  BigHappyButton,
  BigSkyButton,
  GiantBasicButton,
  WarningButton,
} from "../ui/buttons";
import { ArrowRightIcon, CheckIcon } from "@heroicons/react/20/solid";
import Summary from "./summary";
import { PracticeSpotDisplay } from "./practice-spot-display";

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

  const finish = useCallback(
    function (finalSummary: PracticeSummaryItem[]) {
      setSummary(finalSummary);
      setMode("summary");
      document.removeEventListener(
        "UpdateSpotRemindersField",
        updateSpotRemindersField,
      );

      if (pieceid && csrf && startTime) {
        const initialSpots: BasicSpot[] = JSON.parse(initialspots);
        const initialSpotIds = initialSpots.map((spot) => spot.id);
        const spotIDs = finalSummary
          .filter((item) => item.reps > 0)
          .map((item) => item.id)
          .filter((id) => initialSpotIds.includes(id));
        const durationMinutes = Math.ceil(
          (new Date().getTime() - startTime.getTime()) / 1000 / 60,
        );
        document.dispatchEvent(
          new CustomEvent("FinishedSpotPracticing", {
            detail: {
              spotIDs,
              durationMinutes,
              csrf,
              endpoint: `/library/pieces/${pieceid}/practice/random-single`,
            },
          }),
        );
      }
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
        const spots: BasicSpot[] = JSON.parse(initialspots);
        setSpots(spots);
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
  const [practiceSummary, setPracticeSummary] = useState<Map<string, number>>(
    new Map<string, number>(),
  );
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
    function (id: string | undefined) {
      if (!id) {
        return;
      }
      practiceSummary.set(id, (practiceSummary.get(id) ?? 0) + 1);
      setPracticeSummary(practiceSummary);
    },
    [setPracticeSummary, practiceSummary],
  );

  const handleDone = useCallback(
    function () {
      const finalSummary: PracticeSummaryItem[] = [];
      for (const spot of spots) {
        let reps = practiceSummary.get(spot.id) ?? 0;
        if (spots[currentSpotIdx]?.id === spot.id) {
          reps += 1;
        }
        finalSummary.push({
          name: spot.name ?? "Missing spot name",
          reps,
          id: spot.id ?? "Missing spot id",
        });
      }
      finish(finalSummary);
    },
    [practiceSummary, finish, currentSpotIdx, spots],
  );

  const nextSpot = useCallback(
    function () {
      setCounter((curr) => curr + 1);
      addSpotRep(spots[currentSpotIdx]?.id);
      if (skipSpotIds.length >= spots.length) {
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
        (nextSpotId && skipSpotIds.includes(nextSpotId)) ||
        // check if the last two spots are the same and also the same as the next spot
        // but only if the skip spots is more than one smaller than the spots, otherwise
        // there is only one spot left and it will infinitely loop
        (nextSpotId &&
          spots.length - 1 > skipSpotIds.length &&
          lastTwoSpots[0] === lastTwoSpots[1] &&
          lastTwoSpots[1] === nextSpotId)
      ) {
        nextSpotIdx = Math.floor(Math.random() * spots.length);
        nextSpotId = spots[nextSpotIdx]?.id;
      }
      setCurrentSpotIdx(nextSpotIdx);
      setLastTwoSpots([nextSpotId, lastTwoSpots[0]]);
    },
    [addSpotRep, currentSpotIdx, handleDone, skipSpotIds, spots, lastTwoSpots],
  );

  const evictSpot = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      // going to need a copy of this because the it won't be updated by setstate until after the function finishes
      const newSkipSpotIds = [...skipSpotIds];
      if (currentSpotId) {
        newSkipSpotIds.push(currentSpotId);
      }

      if (newSkipSpotIds.length >= spots.length) {
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

      setCounter((curr) => curr + 1);
      addSpotRep(currentSpotId);

      let nextSpotIdx = Math.floor(Math.random() * spots.length);
      let nextSpotId = spots[nextSpotIdx]?.id;
      while (
        !nextSpotId ||
        (nextSpotId && newSkipSpotIds.includes(nextSpotId)) ||
        // check if the last two spots are the same and also the same as the next spot
        // but only if the skip spots is more than one smaller than the spots, otherwise
        // there is only one spot left and it will infinitely loop
        (nextSpotId &&
          spots.length - 1 > newSkipSpotIds.length &&
          lastTwoSpots[0] === lastTwoSpots[1] &&
          lastTwoSpots[1] === nextSpotId)
      ) {
        nextSpotIdx = Math.floor(Math.random() * spots.length);
        nextSpotId = spots[nextSpotIdx]?.id;
      }

      setSkipSpotIds(newSkipSpotIds);

      setCurrentSpotIdx(nextSpotIdx);
      setLastTwoSpots([nextSpotId, lastTwoSpots[0]]);
    },
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, handleDone],
  );

  return (
    <div className="relative w-full" ref={topRef}>
      <div className="absolute left-0 top-0 py-2 sm:py-4">
        <BasicButton onClick={setup}>← Back to setup</BasicButton>
      </div>
      <div className="h-12" />
      <div className="flex w-full flex-col items-center justify-center gap-2 pt-8 sm:pt-24">
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
        <div className="flex flex-col items-center justify-center gap-4 pt-12">
          <div className="flex flex-col justify-center gap-2 md:flex-row">
            <BigHappyButton type="button" onClick={evictSpot}>
              <CheckIcon className="-ml-1 size-6" />
              Finish Spot
            </BigHappyButton>
            <BigSkyButton type="button" onClick={nextSpot}>
              Next Spot <ArrowRightIcon className="-mr-1 size-6" />
            </BigSkyButton>
          </div>
          <p className="mx-auto max-w-2xl text-sm text-neutral-800">
            Once you feel good about a particular spot, you can click “Finish
            Spot” to remove it and keep practicing the others.{" "}
          </p>
        </div>
        <div className="pt-8">
          <WarningButton onClick={handleDone}>Done</WarningButton>
        </div>
      </div>
    </div>
  );
}
