import { BasicSpot, PracticeSummaryItem, RandomMode } from "../common";
import { AnimatePresence, motion } from "framer-motion";
import { StateUpdater, useCallback, useEffect, useState } from "preact/hooks";
import {
  GiantBasicButton,
  BasicButton,
  HappyButton,
  WarningButton,
} from "../ui/buttons";
import { ScaleCrossFadeContent } from "../ui/transitions";
import { CreateSpots } from "./create-spots";
import Summary from "./summary";
import { RadioBox } from "../ui/form";
import { CheckCircleIcon as CheckCircleSolid } from "@heroicons/react/20/solid";
import { CheckCircleIcon as CheckCircleOutline } from "@heroicons/react/24/outline";

export function SequenceSpots({
  initialspots,
  pieceid,
  csrf,
}: {
  initialspots?: string;
  pieceid?: string;
  csrf?: string;
}) {
  const [spots, setSpots] = useState<BasicSpot[]>([]);
  const [summary, setSummary] = useState<PracticeSummaryItem[]>([]);
  const [mode, setMode] = useState<RandomMode>("setup");
  const [fullyRandom, setFullyRandom] = useState(false);
  const [randomizedSpots, setRandomizedSpots] = useState<BasicSpot[]>([]);
  const [startTime, setStartTime] = useState<Date | null>(null);

  useEffect(
    function () {
      if (initialspots) {
        const spots: BasicSpot[] = JSON.parse(initialspots);
        setSpots(spots);
      }
    },
    [initialspots],
  );

  const submit = useCallback(
    function () {
      if (fullyRandom) {
        const randomSpots: BasicSpot[] = [];
        for (let i = 0; i < 2 * spots.length; i++) {
          const randomSpot = spots[Math.floor(Math.random() * spots.length)];
          if (!randomSpot) {
            throw new Error("invalid random spot");
          }
          randomSpots.push(randomSpot);
        }
        setRandomizedSpots(randomSpots);
      } else {
        const tmpSpots = spots.slice();
        tmpSpots.sort(() => Math.random() - 0.5);
        setRandomizedSpots(tmpSpots);
      }
      setMode("practice");
      setStartTime(new Date());
    },
    [fullyRandom, spots, setMode, setRandomizedSpots, setStartTime],
  );

  const moreSpots = useCallback(
    function () {
      if (!fullyRandom) {
        return;
      }
      const randomSpots: BasicSpot[] = [];
      for (let i = 0; i < 2 * spots.length; i++) {
        const randomSpot = spots[Math.floor(Math.random() * spots.length)];
        if (!randomSpot) {
          throw new Error("invalid random spot");
        }
        randomSpots.push(randomSpot);
      }
      setRandomizedSpots((spots) => spots.concat(randomSpots));
    },
    [fullyRandom, spots, setRandomizedSpots],
  );

  const finish = useCallback(
    function (finalSummary: PracticeSummaryItem[]) {
      setSummary(finalSummary);
      setMode("summary");

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
              pieceid,
              csrf,
            },
          }),
        );
      }
    },
    [setMode, setSummary, startTime, pieceid, initialspots, csrf],
  );

  return (
    <div className="w-full">
      <ScaleCrossFadeContent
        component={
          {
            setup: (
              <SequenceSetupForm
                spots={spots}
                setSpots={setSpots}
                fullyRandom={fullyRandom}
                setFullyRandom={setFullyRandom}
                submit={submit}
              />
            ),
            practice: (
              <SequencePractice
                spots={randomizedSpots}
                setup={() => setMode("setup")}
                finish={finish}
                moreSpots={moreSpots}
                fullyRandom={fullyRandom}
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

function SequenceSetupForm({
  setSpots,
  spots,
  fullyRandom,
  setFullyRandom,
  submit,
}: {
  setSpots: StateUpdater<BasicSpot[]>;
  spots: BasicSpot[];
  fullyRandom: boolean;
  setFullyRandom: (fullyRandom: boolean) => void;
  submit: () => void;
}) {
  // TODO: add ability to rate spots

  return (
    <>
      <div className="py-4">
        <h1 className="py-1 text-2xl font-bold">Random Spots Sequence</h1>
        <p className="text-left text-base">
          Add some spots then choose whether to make them completely random, or
          practice each spot one time.
        </p>
      </div>
      <div className="flex flex-col gap-2">
        <div className="flex flex-col">
          <CreateSpots setSpots={setSpots} spots={spots} />
        </div>
        <div className="flex flex-col">
          <label className="text-lg font-semibold text-neutral-800">
            Randomness
          </label>
          <p className="pb-2 text-sm text-neutral-700">
            Choose to either generate a fully random sequence or use each spot
            once before repeating.
          </p>
          <div className="flex w-full gap-2">
            {/* TODO: fix ring, border outline, add no*/}
            <RadioBox
              text="Fully Random"
              setSelected={() => setFullyRandom(true)}
              selected={fullyRandom}
              value="random"
              name="random"
            />
            <RadioBox
              text="Each Spot Once"
              setSelected={() => setFullyRandom(false)}
              selected={!fullyRandom}
              value="each"
              name="each"
            />
          </div>
        </div>
        <div className="col-span-full my-16 flex w-full items-center justify-center">
          <GiantBasicButton disabled={spots.length === 0} onClick={submit}>
            Start Practicing
          </GiantBasicButton>
        </div>
      </div>
    </>
  );
}

function SequencePractice({
  spots,
  setup,
  finish,
  moreSpots,
  fullyRandom,
}: {
  spots: BasicSpot[];
  setup: () => void;
  moreSpots: () => void;
  finish: (summary: PracticeSummaryItem[]) => void;
  fullyRandom: boolean;
}) {
  function handleDone() {
    const finalSummary: PracticeSummaryItem[] = [];
    for (const spot of spots) {
      const spotIndex = finalSummary.findIndex((item) => item.id === spot.id);
      if (spotIndex === -1) {
        finalSummary.push({
          name: spot.name ?? "Missing spot name",
          reps: 1,
          id: spot.id,
        });
      } else {
        if (!finalSummary[spotIndex]) {
          throw new Error("Invalid final summary item");
        }
        finalSummary[spotIndex]!.reps += 1;
      }
    }
    finish(finalSummary);
  }

  return (
    <div className="relative mb-8 grid grid-cols-1">
      <div className="absolute left-0 top-0 sm:py-4">
        <BasicButton onClick={setup}>← Back to setup</BasicButton>
      </div>
      <div className="h-12" />
      <div className="flex w-full flex-col items-center justify-center gap-2 pt-8 sm:pt-24">
        <div className="text-2xl font-semibold text-neutral-700">
          Practicing:
        </div>
        <div className="w-full">
          <ul className="mx-auto grid grid-cols-1 gap-x-4 gap-y-2 sm:grid-cols-2 lg:grid-cols-3">
            {spots.map((spot, idx) => (
              <SpotItem key={`${spot.name}-${idx}`} spot={spot} />
            ))}

            {fullyRandom && (
              <HappyButton onClick={moreSpots}>More Spots...</HappyButton>
            )}
          </ul>
        </div>
      </div>
      <div className="flex w-full items-center justify-center pt-8">
        <WarningButton className="px-6 py-3 text-xl" onClick={handleDone}>
          Done
        </WarningButton>
      </div>
    </div>
  );
}

const variants = {
  initial: {
    scale: 1.1,
  },
  animate: {
    scale: 1,
    transition: { bounce: 0, duration: 0.1 },
  },
  exit: {
    scale: 1.1,
    transition: { duration: 0.1, bounce: 0 },
  },
};

function SpotItem({ spot }: { spot: BasicSpot }) {
  const [completed, setCompleted] = useState(false);
  return (
    <li className="w-full">
      <button
        type="button"
        className="flex w-full justify-between rounded-xl border border-neutral-500 bg-neutral-400/10 p-3 text-xl text-black shadow transition duration-100"
        onClick={() => setCompleted((completed) => !completed)}
      >
        <span className={`${completed ? "font-medium" : "font-base"}`}>
          {spot.name}
        </span>
        <AnimatePresence initial={false} mode="wait">
          {completed ? (
            <motion.span
              key={`${spot.id}-completed`}
              initial="initial"
              animate="animate"
              exit="exit"
              variants={variants}
            >
              <span className="sr-only">Completed</span>
              <CheckCircleSolid className="h-6 w-6 text-green-500" />
            </motion.span>
          ) : (
            <motion.span
              key={`${spot.id}-incomplete`}
              initial="initial"
              animate="animate"
              exit="exit"
              variants={variants}
            >
              <span className="sr-only">Incomplete</span>
              <CheckCircleOutline className="h-6 w-6 text-neutral-500" />
            </motion.span>
          )}
        </AnimatePresence>
      </button>
    </li>
  );
}
