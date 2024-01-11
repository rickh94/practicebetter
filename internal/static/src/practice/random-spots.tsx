import {
  Ref,
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
  VioletButton,
  WarningButton,
} from "../ui/buttons";
import Summary from "./summary";
import { PracticeSpotDisplay } from "./practice-spot-display";
import dayjs from "dayjs";
import { BreakDialog, ResumeDialog } from "./practice-dialogs";

export function RandomSpots({
  initialspots,
  pieceid,
  csrf,
  planid,
  piecetitle,
}: {
  initialspots?: string;
  pieceid?: string;
  csrf?: string;
  planid?: string;
  piecetitle?: string;
}) {
  const [spots, setSpots] = useState<BasicSpot[]>([]);
  const [summary, setSummary] = useState<PracticeSummaryItem[]>([]);
  const [mode, setMode] = useState<RandomMode>("setup");
  const [startTime, setStartTime] = useState<Date | null>(null);
  const [numSessions, setNumSessions] = useState(2);
  const [showPrepare, setShowPrepare] = useState(false);
  const getReadyRef = useRef<HTMLDialogElement>(null);

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
      // get initial sessions value from query param
      const urlParams = new URLSearchParams(window.location.search);
      const initialSessions = parseInt(urlParams.get("numSessions"));
      if (!isNaN(initialSessions) && typeof initialSessions === "number") {
        setNumSessions(initialSessions);
      }
      if (initialspots) {
        const initSpots: BasicSpot[] = JSON.parse(initialspots);
        setSpots(initSpots);
      }
      const skipSetup = !!urlParams.get("skipSetup");
      if (skipSetup) {
        setShowPrepare(true);
        startPracticing();
      }
    },
    [
      initialspots,
      setNumSessions,
      startPracticing,
      window.location.search,
      getReadyRef.current,
    ],
  );

  return (
    <div className="w-full">
      <GetReadyDialog title={piecetitle} show={showPrepare} />
      <ScaleCrossFadeContent
        component={
          {
            setup: (
              <SingleSetupForm
                setSpots={setSpots}
                spots={spots}
                submit={startPracticing}
                numSessions={numSessions}
                setNumSessions={setNumSessions}
              />
            ),
            practice: (
              <SinglePractice
                spots={spots}
                setup={backToSetup}
                finish={finish}
                pieceid={pieceid}
                numSessions={numSessions}
                planid={planid}
              />
            ),
            summary: (
              <Summary
                summary={summary}
                setup={() => setMode("setup")}
                practice={() => setMode("practice")}
                initialSpotIds={initialSpotIds}
                pieceid={pieceid}
                planid={planid}
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

export function GetReadyDialog({
  title,
  show = false,
}: {
  title?: string;
  show?: boolean;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [timeElapsed, setTimeElapsed] = useState(0);
  let interval;
  const closeDialog = useCallback(
    function () {
      globalThis.handleCloseModal();
      if (dialogRef.current) {
        clearInterval(interval);
        dialogRef.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (dialogRef.current) {
              dialogRef.current.classList.remove("close");
              dialogRef.current.close();
            }
          });
        });
      }
    },
    [dialogRef.current],
  );

  useEffect(() => {
    if (show && dialogRef.current) {
      globalThis.handleShowModal();
      dialogRef.current.showModal();
      interval = setInterval(() => {
        console.log("timeElapsed", timeElapsed);
        setTimeElapsed((timeElapsed) => timeElapsed + 1);
      }, 10);
      return () => clearInterval(interval);
    }
  }, [dialogRef.current, setTimeElapsed]);

  useEffect(() => {
    if (timeElapsed >= 200) {
      closeDialog();
    }
  }, [timeElapsed, closeDialog]);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby="resume-title"
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
    >
      <header className="mt-2 text-center sm:text-left">
        <h3
          id="resume-title"
          className="text-2xl font-semibold leading-6 text-neutral-900"
        >
          Random Practicing {title}
        </h3>
      </header>
      <div className="prose prose-neutral mt-2 text-left">
        You’re random practicing{" "}
        <em className="font-medium italic">{title ?? "your spots"}</em>. Get
        ready to start.
      </div>
      <div className="w-full">
        <progress
          id="countdown"
          value={timeElapsed}
          max="200"
          aria-hidden="true"
          className="progress-rounded progress-violet-600 progress-bg-white m-0 w-full"
        >
          {timeElapsed}/200
        </progress>
      </div>
      <div className="mt-2 flex w-full flex-row-reverse flex-wrap gap-2 sm:gap-2">
        <VioletButton grow onClick={closeDialog} className="text-lg">
          Start Practicing
          <span
            className="icon-[iconamoon--player-play-thin] -ml-1 size-5"
            aria-hidden="true"
          ></span>
        </VioletButton>
      </div>
    </dialog>
  );
}

function SingleSetupForm({
  setSpots,
  spots,
  submit,
  numSessions,
  setNumSessions,
}: {
  setSpots: StateUpdater<BasicSpot[]>;
  submit: () => void;
  spots: BasicSpot[];
  numSessions: number;
  setNumSessions: StateUpdater<number>;
}) {
  const numSessionsRef = useRef<HTMLInputElement>(null);
  const handleSubmit = useCallback(() => {
    setNumSessions(parseInt(numSessionsRef.current.value));
    if (spots.length === 0) {
      if (!numSessionsRef.current) {
        setNumSessions(1);
      } else {
        setNumSessions(parseInt(numSessionsRef.current.value));
      }
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
  }, [submit, spots, numSessionsRef.current]);

  const increaseSessions = useCallback(
    function () {
      const curr = parseInt(numSessionsRef.current?.value);
      if (isNaN(curr) || curr < 1) {
        numSessionsRef.current.value = "1";
      }
      numSessionsRef.current.value = (curr + 1).toString();
    },
    [numSessionsRef.current],
  );

  const decreaseSessions = useCallback(
    function () {
      const curr = parseInt(numSessionsRef.current?.value);
      if (isNaN(curr) || curr < 1) {
        numSessionsRef.current.value = "1";
      }
      numSessionsRef.current.value = (curr - 1).toString();
    },
    [numSessionsRef.current],
  );

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
          <div className="flex flex-col gap-2 xs:flex-row">
            <BasicButton onClick={decreaseSessions}>
              <span
                className="icon-[iconamoon--sign-minus-circle-thin] -ml-1 size-5"
                aria-hidden="true"
              />
              Decrease
            </BasicButton>
            <input
              id="num-sessions"
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20 xs:w-20"
              type="number"
              min="1"
              defaultValue={`${numSessions}`}
              ref={numSessionsRef}
            />
            <BasicButton onClick={increaseSessions}>
              <span
                className="icon-[iconamoon--sign-plus-circle-thin] -ml-1 size-5"
                aria-hidden="true"
              />
              Increase
            </BasicButton>
          </div>
        </div>
        <div className="col-span-full my-8 flex w-full items-center justify-center">
          <GiantBasicButton onClick={handleSubmit}>
            Start Practicing
            <span className="icon-[iconamoon--player-play-thin] size-8" />
          </GiantBasicButton>
        </div>
      </div>
    </>
  );
}

// TODO: add events to keep you from moving on while the reminders form is open

function hash(s: string) {
  let h: number, i: number;
  for (i = 0, h = 0xdeadbeef; i < s.length; i++)
    h = Math.imul(h ^ s.charCodeAt(i), 2654435761);
  return (h ^ (h >>> 16)) >>> 0;
}

type SpotNeglectInfo = {
  reps: number;
  evicted: boolean;
};

function findNeglectedSpot(spots: SpotNeglectInfo[]): [boolean, number] {
  let maxReps = -Infinity;
  let minReps = Infinity;
  let minIdx: number = 0;
  for (let i = 0; i < spots.length; i++) {
    if (spots[i].evicted) {
      continue;
    }
    if (spots[i].reps > maxReps) {
      maxReps = spots[i].reps;
    }
    if (spots[i].reps < minReps) {
      minReps = spots[i].reps;
      minIdx = i;
    }
  }

  if (maxReps - minReps > 3) {
    return [true, minIdx];
  } else {
    return [false, -1];
  }
}

// TODO: This should probably be a few components at this point

function SinglePractice({
  spots,
  setup,
  finish,
  pieceid,
  numSessions,
  planid,
}: {
  spots: BasicSpot[];
  setup: () => void;
  finish: (summary: PracticeSummaryItem[]) => void;
  pieceid?: string;
  numSessions: number;
  planid?: string;
}) {
  const [currentSpotIdx, setCurrentSpotIdx] = useState(
    Math.floor(Math.random() * spots.length),
  );
  const [neglectInfo, setNeglectInfo] = useState<SpotNeglectInfo[]>(
    new Array(spots.length).fill({ reps: 0, evicted: false }),
  );
  const [practiceSummary, setPracticeSummary] = useState<
    Map<string, { excellent: number; fine: number; poor: number }>
  >(new Map());
  // This counter ensures that the animation runs, even if the same spot is generated twice in a row.
  const [counter, setCounter] = useState(0);
  const [skipSpotIds, setSkipSpotIds] = useState<string[]>([]);
  const [lastTwoSpots, setLastTwoSpots] = useState<string[]>([]);

  const [sessionsCompleted, setSessionsCompleted] = useState(0);
  const [sessionStarted, setSessionStarted] = useState(dayjs());
  const [hasShownResume, setHasShownResume] = useState(false);
  const [canContinue, setCanContinue] = useState(false);

  const resumeRef = useRef<HTMLDialogElement>(null);
  const topRef = useRef<HTMLDivElement>(null);
  const breakDialogRef = useRef<HTMLDialogElement>(null);
  const spotIdsHash = useMemo(
    function () {
      const ids = spots
        .map((spot) => spot.id)
        .sort()
        .join("");
      return hash(ids);
    },
    [spots],
  );

  const saveToStorage = useCallback(
    function (key: string, value: string) {
      localStorage.setItem(`${pieceid}.${spotIdsHash}.${key}`, value);
      localStorage.setItem(
        `${pieceid}.${spotIdsHash}.savedAt`,
        Date.now().toString(),
      );
    },
    [pieceid, spotIdsHash],
  );

  const loadFromStorage = useCallback(
    function (key: string) {
      const savedAt = localStorage.getItem(`${pieceid}.${spotIdsHash}.savedAt`);
      if (
        !savedAt ||
        dayjs(parseInt(savedAt)).isBefore(dayjs().subtract(1, "day"))
      ) {
        localStorage.removeItem(`${pieceid}.${spotIdsHash}.${key}`);
        return undefined;
      }
      return localStorage.getItem(`${pieceid}.${spotIdsHash}.${key}`);
    },
    [pieceid, spotIdsHash],
  );

  useEffect(() => {
    if (topRef.current) {
      window.scrollTo(0, topRef.current.offsetTop);
    }
  }, [topRef.current]);

  useEffect(() => {
    setHasShownResume(true);
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
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.practiceSummary`);
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.skipSpotIds`);
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.savedAt`);
      return;
    }
  }, [spotIdsHash, pieceid, resumeRef.current, hasShownResume]);

  const handleResume = useCallback(
    function () {
      const summary = loadFromStorage("practiceSummary");
      if (summary) {
        setPracticeSummary(new Map(JSON.parse(summary)));
      }
      const skipSpotIds = loadFromStorage("skipSpotIds");
      if (skipSpotIds) {
        setSkipSpotIds(JSON.parse(skipSpotIds));
        nextSpot(JSON.parse(skipSpotIds));
      } else {
        nextSpot([]);
      }
      const sessionsCompleted = loadFromStorage("sessionsCompleted");
      if (sessionsCompleted) {
        setSessionsCompleted(parseInt(sessionsCompleted));
      }
    },
    [setSkipSpotIds, setPracticeSummary],
  );

  const addSpotRep = useCallback(
    function (id: string | undefined, quality: "excellent" | "fine" | "poor") {
      setNeglectInfo((curr) => {
        curr[currentSpotIdx].reps++;
        return curr;
      });
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
      if (pieceid) {
        saveToStorage(
          "practiceSummary",
          JSON.stringify(Array.from(practiceSummary.entries())),
        );
      }
      setPracticeSummary(practiceSummary);
      return summary;
    },
    [setPracticeSummary, practiceSummary, pieceid, spotIdsHash, currentSpotIdx],
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
      if (pieceid) {
        localStorage.removeItem(`${pieceid}.${spotIdsHash}.practiceSummary`);
        localStorage.removeItem(`${pieceid}.${spotIdsHash}.skipSpotIds`);
        localStorage.removeItem(`${pieceid}.${spotIdsHash}.sessionsCompleted`);
        localStorage.removeItem(`${pieceid}.${spotIdsHash}.savedAt`);
      }
      finish(finalSummary);
    },
    [practiceSummary, finish, spots, pieceid, spotIdsHash],
  );

  const nextSpot = useCallback(
    function (nextSkipSpotIds: string[]) {
      setCounter((curr) => curr + 1);
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
      const [hasNeglectedSpot, neglectedSpotIdx] =
        findNeglectedSpot(neglectInfo);
      let nextSpotIdx: number;
      let nextSpotId: string;
      if (hasNeglectedSpot) {
        nextSpotIdx = neglectedSpotIdx;
      } else {
        nextSpotIdx = Math.floor(Math.random() * spots.length);
        nextSpotId = spots[nextSpotIdx]?.id;
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
        }
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
      if (pieceid) {
        saveToStorage("skipSpotIds", JSON.stringify(newSkipSpotIds));
      }
      setNeglectInfo((curr) => {
        curr[currentSpotIdx].evicted = true;
        return curr;
      });
      setSkipSpotIds(newSkipSpotIds);
      return newSkipSpotIds;
    },
    [spots, skipSpotIds, pieceid, spotIdsHash, currentSpotIdx],
  );

  const maybeTakeABreak = useCallback(
    function () {
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
    },
    [sessionStarted, sessionsCompleted, setSessionsCompleted],
  );

  const handleExcellent = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      const summary = addSpotRep(currentSpotId, "excellent");
      let nextSkipSpotIds = skipSpotIds;
      if (summary.excellent - summary.poor > 4) {
        nextSkipSpotIds = evictSpot(currentSpotId);
      }
      maybeTakeABreak();
      nextSpot(nextSkipSpotIds);
    },
    [
      spots,
      currentSpotIdx,
      skipSpotIds,
      addSpotRep,
      nextSpot,
      evictSpot,
      maybeTakeABreak,
    ],
  );

  const handleFine = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      addSpotRep(currentSpotId, "fine");
      maybeTakeABreak();
      nextSpot(skipSpotIds);
    },
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, nextSpot, maybeTakeABreak],
  );

  const handlePoor = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      const summary = addSpotRep(currentSpotId, "poor");
      let nextSkipSpotIds = skipSpotIds;
      if (summary.poor > 2) {
        nextSkipSpotIds = evictSpot(currentSpotId);
      }
      maybeTakeABreak();
      nextSpot(nextSkipSpotIds);
    },
    [
      spots,
      currentSpotIdx,
      skipSpotIds,
      addSpotRep,
      nextSpot,
      evictSpot,
      maybeTakeABreak,
    ],
  );

  const startSession = useCallback(
    function () {
      setSessionStarted(dayjs());
    },
    [setSessionStarted],
  );

  const takeABreak = useCallback(
    function () {
      if (breakDialogRef.current) {
        breakDialogRef.current.showModal();
        globalThis.handleShowModal();
        setCanContinue(false);
        setTimeout(function () {
          setCanContinue(true);
        }, 30000);
        // }, 1000);
      }
    },
    [breakDialogRef.current, setCanContinue],
  );

  // TODO: handle pretty widths for buttons
  return (
    <div className="relative w-full" ref={topRef}>
      <BreakDialog
        dialogRef={breakDialogRef}
        canContinue={canContinue}
        onContinue={startSession}
        onDone={handleDone}
        planid={planid}
      />
      <ResumeDialog dialogRef={resumeRef} onResume={handleResume} />

      <div className="flex w-full flex-col items-center justify-center gap-2 pt-2">
        <h3 className="text-2xl font-semibold text-neutral-700 underline">
          Random Practicing
        </h3>
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
        <div className="flex w-full flex-col justify-center gap-4 px-4 pt-8 xs:flex-row-reverse xs:px-0">
          <BigHappyButton type="button" onClick={handleExcellent}>
            <span
              className="icon-[iconamoon--like-thin] -ml-1 -mt-1 size-8"
              aria-hidden="true"
            ></span>
            Excellent
          </BigHappyButton>
          <BigSkyButton type="button" onClick={handleFine}>
            <span
              className="icon-[iconamoon--sign-minus-thin] -ml-1 size-8"
              aria-hidden="true"
            ></span>
            Fine
          </BigSkyButton>
          <BigAngryButton type="button" onClick={handlePoor}>
            <span
              className="icon-[iconamoon--dislike-thin] -mb-1 -ml-1 size-8"
              aria-hidden="true"
            ></span>
            Poor
          </BigAngryButton>
        </div>
        <div className="flex justify-center gap-4 pb-12 pt-8">
          <BasicButton onClick={setup}>
            <span
              className="icon-[iconamoon--settings-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>{" "}
            Back to setup
          </BasicButton>
          <WarningButton grow onClick={handleDone}>
            <span
              className="icon-[iconamoon--player-stop-thin] -ml-1 size-5"
              aria-hidden="true"
            ></span>
            Finish
          </WarningButton>
        </div>
      </div>
    </div>
  );
}
