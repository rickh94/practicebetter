import { ScaleCrossFadeContent } from "../ui/transitions";
import {
  BasicButton,
  GiantBasicButton,
  GiantHappyButton,
  WarningButton,
} from "../ui/buttons";
import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { cn, uniqueID } from "../common";
import dayjs from "dayjs";
import { BreakDialog, ResumeDialog } from "./practice-dialogs";
import { SummaryActions } from "./summary";

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
  pieceid = "",
  csrf = "",
  planid = "",
}: {
  initialmeasures?: string;
  initialbeats?: string;
  pieceid?: string;
  csrf?: string;
  planid?: string;
}) {
  const [measures, setMeasures] = useState<number>(
    parseInt(initialmeasures, 10),
  );
  const [beats, setBeats] = useState<number>(parseInt(initialbeats, 10));
  const [mode, setMode] = useState<StartingPointMode>("setup");
  const [maxLength, setMaxLength] = useState<number>(5);
  const [summary, setSummary] = useState<Section[]>([]);
  const [measuresPracticed, setMeasuresPracticed] = useState<
    [number, number][]
  >([]);
  const [startTime, setStartTime] = useState<Date | null>(null);
  const [numSessions, setNumSessions] = useState(2);

  const [lowerBound, setLowerBound] = useState<number | null>(null);
  const [upperBound, setUpperBound] = useState<number | null>(null);

  const setModePractice = useCallback(() => {
    setSummary([]);
    setMode("practice");
    setStartTime(new Date());
  }, [setMode, setSummary, setStartTime]);

  const saveConfig = useCallback(
    (c: FormData) => {
      const measures = c.get("measures");
      if (measures && typeof measures === "string") {
        const measuresInt = parseInt(measures, 10);
        if (!isNaN(measuresInt)) {
          setMeasures(measuresInt);
        }
      }

      const beats = c.get("beats");
      if (beats && typeof beats === "string") {
        const beatsInt = parseInt(beats, 10);
        if (!isNaN(beatsInt)) {
          setBeats(beatsInt);
        }
      }

      const maxLength = c.get("maxLength");
      if (maxLength && typeof maxLength === "string") {
        const maxLengthInt = parseInt(maxLength, 10);
        if (!isNaN(maxLengthInt)) {
          setMaxLength(maxLengthInt);
        }
      }

      const numSessions = c.get("numSessions");
      if (numSessions && typeof numSessions === "string") {
        const numSessionsInt = parseInt(numSessions, 10);
        if (!isNaN(numSessionsInt)) {
          setNumSessions(numSessionsInt);
        }
      }

      const lowerBound = c.get("lowerBound");
      if (lowerBound && typeof lowerBound === "string") {
        const lowerBoundInt = parseInt(lowerBound, 10);
        if (!isNaN(lowerBoundInt)) {
          setLowerBound(lowerBoundInt);
        }
      }

      const upperBound = c.get("upperBound");
      if (upperBound && typeof upperBound === "string") {
        const upperBoundInt = parseInt(upperBound, 10);
        if (!isNaN(upperBoundInt)) {
          setUpperBound(upperBoundInt);
        }
      }
      setModePractice();
    },
    [
      setBeats,
      setMeasures,
      setMaxLength,
      setNumSessions,
      setLowerBound,
      setUpperBound,
      setModePractice,
    ],
  );

  const setModeSetup = useCallback(() => {
    setSummary([]);
    setMode("setup");
  }, [setMode, setSummary]);

  const finishPracticing = useCallback(
    (finalSummary: Section[]) => {
      setMode("summary");
      setSummary(finalSummary);
      const mpracticed = calculateMeasuresPracticed(finalSummary);
      setMeasuresPracticed(mpracticed);

      if (pieceid && csrf && startTime) {
        const durationMinutes = Math.ceil(
          (new Date().getTime() - startTime.getTime()) / 1000 / 60,
        );

        const mp = mpracticed
          .map(([start, end]) =>
            start === end ? `${start}` : `${start}-${end}`,
          )
          .join(", ");
        globalThis.dispatchEvent(
          new CustomEvent("FinishedStartingPointPracticing", {
            detail: {
              measuresPracticed: mp,
              durationMinutes,
              pieceid,
              csrf,
            },
          }),
        );
      }
    },
    [setSummary, setMode, pieceid, setMeasuresPracticed, startTime, csrf],
  );

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const skipSetup = !!urlParams.get("skipSetup");
    if (skipSetup) {
      setModePractice();
    }
  }, [setModePractice]);

  // TODO: consider switch to react-hook-form for setup to reduce annoying prop complexity
  return (
    <div className="relative left-0 top-0 w-full sm:mx-auto sm:max-w-6xl">
      <ScaleCrossFadeContent
        component={
          {
            setup: (
              <StartingPointSetupForm
                preconfigured={!!pieceid}
                beats={beats}
                measures={measures}
                maxLength={maxLength}
                lowerBound={lowerBound}
                upperBound={upperBound}
                submit={saveConfig}
                numSessions={numSessions}
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
                numSessions={numSessions}
                pieceid={pieceid}
                planid={planid}
              />
            ),
            summary: (
              <Summary
                summary={summary}
                measuresPracticed={measuresPracticed}
                setup={setModeSetup}
                practice={setModePractice}
                pieceid={pieceid}
                planid={planid}
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
  submit,
  numSessions,
}: {
  beats: number;
  measures: number;
  maxLength: number;
  lowerBound: number | null;
  upperBound: number | null;
  preconfigured: boolean;
  numSessions?: number;
  submit: (c: FormData) => void;
}) {
  const upperBoundRef = useRef<HTMLInputElement>(null);
  const lowerBoundRef = useRef<HTMLInputElement>(null);
  const formRef = useRef<HTMLFormElement>(null);

  const autoSelect = useCallback((e: FocusEvent) => {
    if (e.currentTarget instanceof HTMLInputElement) {
      e.currentTarget?.select();
    }
  }, []);

  const handleSubmit = useCallback(
    (e: Event) => {
      e.preventDefault();
      if (!formRef.current) {
        return;
      }
      const data = new FormData(formRef.current);
      submit(data);
    },
    [submit],
  );

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
      <form
        onSubmit={handleSubmit}
        ref={formRef}
        className="grid w-full grid-cols-1 grid-rows-8 gap-1 sm:grid-cols-2 sm:grid-rows-4"
      >
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
            name="measures"
            className={cn(
              "focusable w-24 rounded-xl px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20",
              preconfigured
                ? "bg-black/30 text-neutral-500"
                : "bg-neutral-700/10",
            )}
            type="number"
            min="2"
            defaultValue={`${measures}`}
            disabled={preconfigured}
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
              "focusable w-24 rounded-xl px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20",
              preconfigured
                ? "bg-black/30 text-neutral-500"
                : "bg-neutral-700/10",
            )}
            type="number"
            min="1"
            name="beats"
            defaultValue={`${beats}`}
            disabled={preconfigured}
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
              name="maxLength"
              className="focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              defaultValue={`${maxLength}`}
              onFocus={autoSelect}
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
              name="lowerBound"
              className="focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-400 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              placeholder="mm"
              defaultValue={`${lowerBound ?? ""}`}
              onFocus={autoSelect}
              ref={lowerBoundRef}
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
              name="upperBound"
              className="focusable w-24 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-400 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              max={measures}
              placeholder="mm"
              value={`${upperBound ?? ""}`}
              onFocus={autoSelect}
              ref={upperBoundRef}
            />
          </div>
          <div className="flex h-full flex-col justify-end gap-1 sm:pb-2">
            <BasicButton
              onClick={() => {
                if (lowerBoundRef.current) {
                  lowerBoundRef.current.value = "";
                }
                if (upperBoundRef.current) {
                  upperBoundRef.current.value = "";
                }
              }}
              type="button"
            >
              <span
                className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-5"
                aria-hidden="true"
              />
              Clear
            </BasicButton>
          </div>
        </div>

        <div className="flex flex-col">
          <label
            className="text-lg font-semibold text-neutral-800"
            for="num-sessions"
          >
            Number of Sessions
          </label>
          <p className="pb-2 text-sm text-neutral-700">
            Random practicing is broken up into five minute sessions with one
            minute breaks.
          </p>
          <div className="flex gap-2">
            <input
              id="num-sessions"
              name="numSessions"
              className="focusable w-20 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              defaultValue={`${numSessions}`}
            />
          </div>
        </div>
        <div className="col-span-full my-8 flex w-full items-center justify-center">
          <GiantBasicButton type="submit">
            Start Practicing
            <span className="icon-[iconamoon--player-play-thin] size-8" />
          </GiantBasicButton>
        </div>
      </form>
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
  pieceid = "",
  numSessions = 2,
  planid = "",
}: {
  beats: number;
  measures: number;
  maxLength: number;
  lowerBound: number | null;
  upperBound: number | null;
  pieceid?: string;
  setup: () => void;
  finish: (summary: Section[]) => void;
  planid?: string;
  numSessions?: number;
}) {
  const [practiceSummary, setPracticeSummary] = useState<Section[]>([]);
  const [section, setSection] = useState<Section>(
    makeRandomSection(measures, beats, maxLength, lowerBound, upperBound),
  );
  const topRef = useRef<HTMLDivElement>(null);

  const [sessionsCompleted, setSessionsCompleted] = useState(0);
  const [sessionStarted, setSessionStarted] = useState(dayjs());
  const [hasShownResume, setHasShownResume] = useState(false);
  const resumeRef = useRef<HTMLDialogElement>(null);
  const breakDialogRef = useRef<HTMLDialogElement>(null);
  const [canContinue, setCanContinue] = useState(false);

  const saveToStorage = useCallback(
    (key: string, value: string) => {
      if (pieceid) {
        localStorage.setItem(`${pieceid}.startingPoint.${key}`, value);
        localStorage.setItem(
          `${pieceid}.startingPoint.savedAt`,
          Date.now().toString(),
        );
      }
    },
    [pieceid],
  );

  const loadFromStorage = useCallback(
    (key: string) => {
      if (pieceid) {
        const savedAt = localStorage.getItem(
          `${pieceid}.startingPoint.savedAt`,
        );
        if (
          !savedAt ||
          dayjs(parseInt(savedAt, 10)).isBefore(dayjs().subtract(1, "day"))
        ) {
          localStorage.removeItem(`${pieceid}.startingPoint.${key}`);
          return undefined;
        }
        return localStorage.getItem(`${pieceid}.startingPoint.${key}`);
      }
    },
    [pieceid],
  );

  useEffect(() => {
    if (topRef.current) {
      window.scrollTo(0, topRef.current.offsetTop);
    }
  }, []);

  const handleResume = useCallback(() => {
    const summary = loadFromStorage("practiceSummary");
    if (summary) {
      setPracticeSummary(JSON.parse(summary) as Section[]);
    }
    const sessionsCompleted = loadFromStorage("sessionsCompleted");
    if (sessionsCompleted) {
      setSessionsCompleted(parseInt(sessionsCompleted, 10));
    }
  }, [loadFromStorage]);

  useEffect(() => {
    setHasShownResume(true);
    if (!pieceid) {
      return;
    }
    const practiceSummary = loadFromStorage("practiceSummary");
    if (practiceSummary) {
      const autoResume = new URLSearchParams(window.location.search).get(
        "resume",
      );
      if (autoResume) {
        handleResume();
        return;
      }
      if (resumeRef.current && !hasShownResume) {
        resumeRef.current.showModal();
        globalThis.handleShowModal();
      }
    } else {
      localStorage.removeItem(`${pieceid}.startingPoint.practiceSummary`);
      localStorage.removeItem(`${pieceid}.startingPoint.sessionsCompleted`);
      localStorage.removeItem(`${pieceid}.startingPoint.savedAt`);
      return;
    }
  }, [pieceid, hasShownResume, loadFromStorage, handleResume]);

  const takeABreak = useCallback(() => {
    if (breakDialogRef.current) {
      breakDialogRef.current.showModal();
      globalThis.handleShowModal();
      setCanContinue(false);
      setTimeout(() => {
        setCanContinue(true);
      }, 45000);
      // }, 1000);
    }
  }, [setCanContinue]);

  const handleDone = useCallback(() => {
    // have to add the last one in manually
    const finalSummary = [...practiceSummary, section];
    finalSummary.sort(
      (a, b) => a.startingPoint.measure - b.startingPoint.measure,
    );
    localStorage.removeItem(`${pieceid}.startingPoint.practiceSummary`);
    localStorage.removeItem(`${pieceid}.startingPoint.sessionsCompleted`);
    localStorage.removeItem(`${pieceid}.startingPoint.savedAt`);
    finish(finalSummary);
  }, [practiceSummary, section, pieceid, finish]);

  const startSession = useCallback(() => {
    setSessionStarted(dayjs());
  }, [setSessionStarted]);

  const maybeTakeABreak = useCallback(() => {
    if (dayjs().diff(sessionStarted, "minute") < 5) {
      // if (dayjs().diff(sessionStarted, "second") < 2) {
      return;
    }
    if (sessionsCompleted >= numSessions - 1) {
      handleDone();
    } else {
      takeABreak();
      setSessionsCompleted((curr) => curr + 1);
      saveToStorage("sessionsCompleted", `${sessionsCompleted + 1}`);
    }
  }, [
    handleDone,
    numSessions,
    saveToStorage,
    sessionStarted,
    sessionsCompleted,
    takeABreak,
  ]);

  const nextStartingPoint = useCallback(() => {
    const nextPracticeSummary = [...practiceSummary, section];
    setPracticeSummary(nextPracticeSummary);
    saveToStorage("practiceSummary", JSON.stringify(nextPracticeSummary));
    setSection(
      makeRandomSection(measures, beats, maxLength, lowerBound, upperBound),
    );
    maybeTakeABreak();
  }, [
    practiceSummary,
    section,
    saveToStorage,
    measures,
    beats,
    maxLength,
    lowerBound,
    upperBound,
    maybeTakeABreak,
  ]);

  return (
    <div className="relative mb-8 grid grid-cols-1" ref={topRef}>
      <BreakDialog
        dialogRef={breakDialogRef}
        canContinue={canContinue}
        onContinue={startSession}
        onDone={handleDone}
        planid={planid}
      />
      <ResumeDialog dialogRef={resumeRef} onResume={handleResume} />
      <div className="absolute left-0 top-0 sm:p-8" />
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
        <div className="flex flex-wrap justify-center gap-2 pt-8">
          <BasicButton onClick={setup} type="button">
            <span
              className="icon-[iconamoon--arrow-left-5-circle-thin] -ml-1 size-5"
              aria-hidden="true"
            />{" "}
            Back to setup
          </BasicButton>
          <WarningButton onClick={handleDone}>
            <span
              className="icon-[iconamoon--player-stop-thin] -ml-1 size-5"
              aria-hidden="true"
            />
            Finish
          </WarningButton>
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
  pieceid,
  planid,
}: {
  summary: Section[];
  measuresPracticed: [number, number][];
  setup: () => void;
  practice: () => void;
  pieceid?: string;
  planid?: string;
}) {
  return (
    <>
      <div className="flex w-full flex-col justify-center gap-4 pb-8 pt-4 sm:flex-row sm:gap-6 sm:pt-8">
        <SummaryActions
          setup={setup}
          practice={practice}
          pieceid={pieceid}
          planid={planid}
        />
      </div>
      <div className="flex w-full flex-col items-center justify-center gap-2">
        <div className="flex w-full  justify-center py-4">
          <div className="rounded-xl border border-neutral-500 bg-white/80 px-6 py-4 text-center shadow">
            <div className="flex w-full justify-center">
              <h2 className="border-b border-black px-2 text-center text-xl font-semibold text-black">
                Measures Practiced
              </h2>
            </div>
            <div className="text-balance pt-1">
              {measuresPracticed.map(([start, end], idx) => (
                <span
                  key={`${start}-${end}`}
                  className="whitespace-nowrap text-xl font-medium text-neutral-800"
                >
                  {start === end ? start : `${start}-${end}`}
                  {idx < measuresPracticed.length - 1 && ","}
                </span>
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
