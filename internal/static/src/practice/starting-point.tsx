import { ScaleCrossFadeContent } from "../ui/transitions";
import { XMarkIcon } from "@heroicons/react/20/solid";
import {
  BasicButton,
  GiantBasicButton,
  GiantHappyButton,
  HappyButton,
  WarningButton,
} from "../ui/buttons";
import { BackToPieceLink } from "../ui/links";
import { StateUpdater, useCallback, useState } from "preact/hooks";
import { cn, uniqueID } from "../common";

type Section = {
  startingPoint: {
    measure: number;
    beat: number;
  };
  endingPoint: {
    measure: number;
    beat: number;
  };
  id: string;
};

type StartingPointMode = "setup" | "practice" | "summary";

function calculateMeasuresPracticed(summary: Section[]) {
  const measureSet = new Set<number>();
  for (const { startingPoint, endingPoint } of summary) {
    for (let i = startingPoint.measure; i <= endingPoint.measure; i++) {
      measureSet.add(i);
    }
  }
  const measureList = Array.from(measureSet.values()).sort((a, b) => a - b);
  const ranges: [number, number][] = [];

  // walk the list, while the numbers are sequential, keep increasing the end number,
  // once we hit a gap, add a range, reset to start at the current number and continue.
  let start: number | null = null;
  let end: number | null = null;
  for (const num of measureList) {
    if (!start || !end) {
      start = num;
      end = num;
    } else if (num === end + 1) {
      end = num;
    } else {
      if (!end) {
        end = start;
      }
      ranges.push([start, end]);
      start = num;
      end = num;
    }
  }
  if (start && end) {
    ranges.push([start, end]);
  }
  return ranges;
}

// TODO: add option for time signature changes
// TODO: add option to focus on smaller section
export function StartingPoint({
  initialmeasures = "100",
  initialbeats = "4",
  preconfigured = false,
  pieceid,
  csrf,
}: {
  initialmeasures?: string;
  initialbeats?: string;
  preconfigured?: boolean;
  pieceid?: string;
  csrf?: string;
}) {
  const [measures, setMeasures] = useState<number>(parseInt(initialmeasures));
  const [beats, setBeats] = useState<number>(parseInt(initialbeats));
  const [mode, setMode] = useState<StartingPointMode>("setup");
  const [maxLength, setMaxLength] = useState<number>(5);
  const [summary, setSummary] = useState<Section[]>([]);
  const [measuresPracticed, setMeasuresPracticed] = useState<
    [number, number][]
  >([]);
  const [startTime, setStartTime] = useState<Date | null>(null);

  const [lowerBound, setLowerBound] = useState<number | null>(null);
  const [upperBound, setUpperBound] = useState<number | null>(null);

  const setModePractice = useCallback(
    function () {
      setSummary([]);
      setMode("practice");
      setStartTime(new Date());
    },
    [setMode],
  );

  const setModeSetup = useCallback(
    function () {
      setSummary([]);
      setMode("setup");
    },
    [setMode],
  );

  const finishPracticing = useCallback(
    function (finalSummary: Section[]) {
      setMode("summary");
      setSummary(finalSummary);
      const mpracticed = calculateMeasuresPracticed(finalSummary);
      setMeasuresPracticed(mpracticed);

      if (pieceid && csrf && startTime) {
        const durationMinutes = Math.ceil(
          (new Date().getTime() - startTime.getTime()) / 1000 / 60,
        );
        document.dispatchEvent(
          new CustomEvent("FinishedStartingPointPracticing", {
            detail: {
              measuresPracticed: mpracticed,
              durationMinutes,
            },
          }),
        );
      }
    },
    [setSummary, setMode, pieceid, setMeasuresPracticed, startTime],
  );

  // TODO: consider switch to react-hook-form for setup to reduce annoying prop complexity
  return (
    <div className="relative left-0 top-0 w-full sm:mx-auto sm:max-w-5xl">
      <ScaleCrossFadeContent
        component={
          {
            setup: (
              <StartingPointSetupForm
                preconfigured={preconfigured}
                beats={beats}
                measures={measures}
                maxLength={maxLength}
                setMaxLength={setMaxLength}
                setBeats={setBeats}
                setMeasures={setMeasures}
                submit={setModePractice}
                lowerBound={lowerBound}
                upperBound={upperBound}
                setLowerBound={setLowerBound}
                setUpperBound={setUpperBound}
              />
            ),
            practice: (
              <StartingPointPractice
                beats={beats}
                measures={measures}
                maxLength={maxLength}
                setup={setModeSetup}
                finish={finishPracticing}
                lowerBound={lowerBound}
                upperBound={upperBound}
              />
            ),
            summary: (
              <Summary
                summary={summary}
                measuresPracticed={measuresPracticed}
                setup={setModeSetup}
                practice={setModePractice}
                pieceHref={pieceHref}
              />
            ),
          }[mode]
        }
        id={mode}
      />
    </div>
  );
}

// TODO: switch to grid layout for better alignment
// TODO: rewrite description
export function StartingPointSetupForm({
  beats,
  measures,
  maxLength,
  lowerBound,
  upperBound,
  preconfigured,
  setMaxLength,
  setBeats,
  setMeasures,
  setLowerBound,
  setUpperBound,
  submit,
}: {
  beats: number;
  measures: number;
  maxLength: number;
  lowerBound: number | null;
  upperBound: number | null;
  preconfigured: boolean;
  setMaxLength: StateUpdater<number>;
  setBeats: StateUpdater<number>;
  setLowerBound: StateUpdater<number | null>;
  setUpperBound: StateUpdater<number | null>;
  setMeasures: StateUpdater<number>;
  submit: () => void;
}) {
  function isValid() {
    return beats > 0 && measures > 0;
  }

  const autoSelect = useCallback(function (e: FocusEvent) {
    if (e.currentTarget instanceof HTMLInputElement) {
      e.currentTarget?.select();
    }
  }, []);

  // TODO: improve form
  return (
    <>
      <div className="z-0 flex w-full flex-col">
        <div>
          <h1 className="py-1 text-left text-2xl font-bold">
            Random Starting Point
          </h1>
          <p className="text-left text-base">
            Enter the number of measures and beats per measure of your piece,
            then you can practice starting and stopping at random spots
          </p>
        </div>
      </div>
      <div className="grid-rows-8 grid w-full grid-cols-1 gap-1 sm:grid-cols-2 sm:grid-rows-4">
        <div className="col-span-1 col-start-1 row-span-1 row-start-1">
          <label
            className="text-lg font-semibold text-neutral-800"
            htmlFor="measures"
          >
            Measures{" "}
            {preconfigured && (
              <span className="rounded-xl bg-black/10 p-1 font-extrabold text-black">
                (set automatically)
              </span>
            )}
          </label>
          <p className="pb-2 text-sm text-neutral-700">
            How many measures are in your piece?
          </p>
        </div>
        <div className="col-span-1 col-start-1 row-span-1 flex items-center gap-2 sm:row-start-2">
          <input
            id="measures"
            className={cn(
              "focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20",
              preconfigured && "bg-black/30 text-neutral-500",
            )}
            type="number"
            min="2"
            value={measures}
            disabled={preconfigured}
            // @ts-ignore
            onChange={(e: InputEvent) => setMeasures(parseInt(e.target.value))}
            onFocus={autoSelect}
          />
          <div className="font-medium">Measures</div>
        </div>
        <div className="col-span-1 row-span-1 sm:col-start-2">
          <label
            className="text-lg font-semibold text-neutral-800"
            htmlFor="beats"
          >
            Beats per measure{" "}
            {preconfigured && (
              <span className="rounded-xl bg-black/10 p-1 font-extrabold text-black">
                (set automatically)
              </span>
            )}
          </label>
          <p className="text-sm text-neutral-700">
            How many beats are in each measure?
          </p>
          <p className="pb-2 text-sm italic text-neutral-700">
            (the top number from the time signature)
          </p>
        </div>
        <div className="col-span-1 row-span-1 flex items-center gap-2 sm:col-start-2 sm:row-start-2">
          <input
            id="beats"
            className={cn(
              "focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20",
              preconfigured && "bg-black/30 text-neutral-500",
            )}
            type="number"
            min="1"
            value={beats}
            disabled={preconfigured}
            // @ts-ignore
            onChange={(e) => setBeats(parseInt(e.target.value))}
            onFocus={autoSelect}
          />
          <div className="font-medium">Beats</div>
        </div>
        <div className="col-span-1 col-start-1 row-span-1 sm:row-start-3">
          <label
            className="text-lg font-semibold text-neutral-800"
            htmlFor="maxLength"
          >
            Maximum Length
          </label>
          <p className="text-sm text-neutral-700">
            The sections will be of random number of measures less than this
            number.
          </p>
        </div>
        <div className="col-span-1 col-start-1 row-span-1 flex items-end gap-2 pb-2 sm:row-start-4">
          <div className="flex items-center gap-2">
            <input
              id="maxLength"
              className="focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              value={maxLength}
              onFocus={autoSelect}
              // @ts-ignore
              onChange={(e) => setMaxLength(parseInt(e.target.value))}
            />
            <div className="font-medium">Measures</div>
          </div>
        </div>
        <div className="col-span-1 col-start-1 row-span-1 sm:col-start-2 sm:row-start-3">
          <div className="text-lg font-semibold text-neutral-800">
            Limit to Measures (optional)
          </div>
          <p className="text-sm text-neutral-700">
            You can limit practicing to a smaller section within the piece.
          </p>
        </div>
        <div className="col-span-1 col-start-1 row-span-1 flex items-center gap-2 sm:col-start-2 sm:row-start-4">
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium text-neutral-800"
              htmlFor="lowerBound"
            >
              Start
            </label>
            <input
              id="lowerBound"
              className="focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-400 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              placeholder="mm"
              value={lowerBound ?? ""}
              onFocus={autoSelect}
              // @ts-ignore
              onChange={(e) => setLowerBound(parseInt(e.target.value))}
            />
          </div>
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium text-neutral-800"
              htmlFor="upperBound"
            >
              End
            </label>
            <input
              id="upperBound"
              className="focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-400 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              max={measures}
              placeholder="mm"
              value={upperBound ?? ""}
              onFocus={autoSelect}
              // @ts-ignore
              onChange={(e) => setUpperBound(parseInt(e.target.value))}
            />
          </div>
          <div className="flex h-full flex-col justify-end gap-1 sm:pb-2">
            <BasicButton
              onClick={() => {
                setLowerBound(null);
                setUpperBound(null);
              }}
              type="button"
            >
              <XMarkIcon className="-ml-1 h-5 w-5" />
              Clear
            </BasicButton>
          </div>
        </div>
        <div className="col-span-full my-16 flex w-full items-center justify-center">
          <GiantBasicButton disabled={!isValid()} onClick={submit}>
            Start Practicing
          </GiantBasicButton>
        </div>
      </div>
    </>
  );
}

// TODO: maybe write a test for this function
export function makeRandomSection(
  measures: number,
  beats: number,
  maxLength: number,
  lowerBound: number | null,
  upperBound: number | null,
): Section {
  // subtract one so we never start in the last measure
  let randomStartingMeasure = Math.floor(Math.random() * (measures - 1)) + 1;
  while (
    (lowerBound !== null && randomStartingMeasure < lowerBound) ||
    (upperBound !== null && randomStartingMeasure > upperBound)
  ) {
    randomStartingMeasure = Math.floor(Math.random() * (measures - 1)) + 1;
  }
  // make sure we don't go past the end
  const maxOffset = Math.min(maxLength, measures - randomStartingMeasure);
  // add back in the measure we subtracted earlier.
  const randomEndingMeasure =
    Math.floor(Math.random() * maxOffset) + randomStartingMeasure + 1;

  return {
    startingPoint: {
      measure: randomStartingMeasure,
      // generates between 0 and beats - 1, so we need to add one
      beat: Math.floor(Math.random() * beats) + 1,
    },
    endingPoint: {
      measure: randomEndingMeasure,
      beat: Math.floor(Math.random() * beats) + 1,
    },
    id: uniqueID(),
  };
}

export function StartingPointPractice({
  beats,
  measures,
  maxLength,
  lowerBound,
  upperBound,
  setup,
  finish,
}: {
  beats: number;
  measures: number;
  maxLength: number;
  lowerBound: number | null;
  upperBound: number | null;
  setup: () => void;
  finish: (summary: Section[]) => void;
}) {
  const [practiceSummary, setPracticeSummary] = useState<Section[]>([]);
  const [section, setSection] = useState<Section>(
    makeRandomSection(measures, beats, maxLength, lowerBound, upperBound),
  );

  const nextStartingPoint = useCallback(
    function () {
      setPracticeSummary((curr) => [...curr, section]);
      setSection(
        makeRandomSection(measures, beats, maxLength, lowerBound, upperBound),
      );
    },
    [beats, measures, section, maxLength, lowerBound, upperBound],
  );

  const handleDone = useCallback(
    function () {
      // have to add the last one in manually
      const finalSummary = [...practiceSummary, section];
      finalSummary.sort(
        (a, b) => a.startingPoint.measure - b.startingPoint.measure,
      );
      finish(finalSummary);
    },
    [practiceSummary, finish, section],
  );

  return (
    <div className="relative mb-8 grid grid-cols-1">
      <div className="absolute left-0 top-0 sm:p-8">
        <BasicButton onClick={setup} type="button">
          ← Back to setup
        </BasicButton>
      </div>
      <div className="h-12" />
      <div className="flex w-full flex-col items-center justify-center gap-2 pt-12 sm:pt-24">
        <div className="relative h-32 w-full">
          <ScaleCrossFadeContent
            component={<SectionDisplay section={section} />}
            id={section.id}
          />
        </div>
        <div className="pt-8">
          <GiantHappyButton onClick={nextStartingPoint}>
            Next Section
          </GiantHappyButton>
        </div>
        <div className="pt-8">
          <WarningButton onClick={handleDone}>Done</WarningButton>
        </div>
      </div>
    </div>
  );
}

export function SectionDisplay({ section }: { section: Section }) {
  return (
    <div className="flex justify-center">
      <div className="rounded-xl border border-neutral-500 bg-white/90 px-4 pb-5 pt-4 text-center text-lg shadow-lg sm:px-8 sm:text-xl">
        <div>
          Start from measure{" "}
          <strong className="text-xl font-bold sm:text-2xl">
            {section.startingPoint.measure}
          </strong>
          {", "}beat{" "}
          <strong className="text-xl font-bold sm:text-2xl">
            {section.startingPoint.beat}
          </strong>{" "}
        </div>
        <div>
          and play until measure{" "}
          <strong className="text-xl font-bold sm:text-2xl">
            {section.endingPoint.measure}
          </strong>
          {", "}
          beat{" "}
          <strong className="text-xl font-bold sm:text-2xl">
            {section.endingPoint.beat}
          </strong>
          .
        </div>
      </div>
    </div>
  );
}

export function Summary({
  summary,
  measuresPracticed,
  setup,
  practice,
  pieceHref,
}: {
  summary: Section[];
  measuresPracticed: [number, number][];
  setup: () => void;
  practice: () => void;
  pieceHref?: string;
}) {
  return (
    <>
      <div className="flex w-full flex-col justify-center gap-4 pb-8 pt-12 sm:flex-row  sm:gap-6">
        {pieceHref && <BackToPieceLink pieceHref={pieceHref} />}
        <WarningButton onClick={setup}>Back to Setup</WarningButton>
        <HappyButton onClick={practice}>Practice More</HappyButton>
      </div>
      <div className="flex w-full flex-col items-center justify-center gap-2">
        <div className="flex w-full  justify-center py-4">
          <div className="rounded-xl border border-neutral-500 bg-white/80 px-6 py-4 text-center shadow">
            <div className="flex w-full justify-center">
              <h2 className="border-b border-black px-2 text-center text-xl font-semibold text-black">
                Measures Practiced
              </h2>
            </div>
            <div className="balanced pt-1">
              {measuresPracticed.map(([start, end], idx) => (
                <>
                  <span
                    key={`${start}-${end}`}
                    className="whitespace-nowrap text-xl font-medium text-neutral-800"
                  >
                    {start === end ? start : `${start}-${end}`}
                    {idx < measuresPracticed.length - 1 && ","}
                  </span>{" "}
                </>
              ))}
            </div>
          </div>
        </div>
        <h2 className="inline pr-2 text-xl font-semibold text-black">
          Section Summary
        </h2>
        <table className="min-w-full divide-y divide-neutral-700">
          <thead>
            <tr>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:pl-0"
              >
                Starting Point
              </th>
              <th
                scope="col"
                className="hidden py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:block sm:pl-0"
              >
                Through
              </th>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:pl-0"
              >
                Ending Point
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-neutral-700 text-sm sm:text-base">
            {summary.map(({ startingPoint, endingPoint, id }, idx) => (
              <tr
                key={id}
                className={`${idx % 2 === 0 && "bg-neutral-700/10"}`}
              >
                <td className="whitespace-nowrap py-2 pl-4 pr-3 text-center font-medium text-neutral-900 sm:pl-0">
                  Measure {startingPoint.measure}, beat {startingPoint.beat}
                </td>
                <td className="hidden whitespace-nowrap py-2 pl-4 pr-3 text-center font-medium text-neutral-900 sm:block sm:pl-0">
                  —
                </td>
                <td className="whitespace-nowrap px-3 py-2 text-center text-neutral-800">
                  Measure {endingPoint.measure}, beat {endingPoint.beat}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
