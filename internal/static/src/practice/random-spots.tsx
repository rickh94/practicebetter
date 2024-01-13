import {
  type StateUpdater,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "preact/hooks";
import { type PracticeSummaryItem, type RandomMode } from "../common";
import { type BasicSpot } from "../validators";
import { ScaleCrossFadeContent } from "../ui/transitions";
import { CreateSpots } from "./create-spots";
import {
  BasicButton,
  BigAngryButton,
  BigHappyButton,
  BigSkyButton,
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
  const [startTime, setStartTime] = useState<Date | undefined>(undefined);
  const [numSessions, setNumSessions] = useState(2);
  const [showPrepare, setShowPrepare] = useState(false);

  const initialSpotIds = useMemo(() => {
    const spots: string[] = [];
    if (!initialspots) return spots;
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const initialSpots: BasicSpot[] = JSON.parse(initialspots);
    if (!(initialSpots instanceof Array) || initialSpots.length === 0) {
      return spots;
    }
    for (const spot of initialSpots) {
      if (spot.id) {
        spots.push(spot.id);
      }
    }
    return spots;
  }, [initialspots]);

  const updateSpotRemindersField = useCallback(
    (event: CustomEvent) => {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
      const { id, text } = event.detail;
      if (!id || !text || typeof id !== "string" || typeof text !== "string") {
        throw new Error("Invalid event");
      }
      setSpots((spots) =>
        spots.map((spot) =>
          spot.id === id ? { ...spot, textPrompt: text } : spot,
        ),
      );
    },
    [setSpots],
  );
  const finish = useCallback(
    (finalSummary: PracticeSummaryItem[]) => {
      setSummary(finalSummary);
      setMode("summary");
      globalThis.removeEventListener(
        "UpdateSpotRemindersField",
        updateSpotRemindersField,
      );
    },
    [updateSpotRemindersField],
  );

  const startPracticing = useCallback(() => {
    setStartTime(new Date());
    setMode("practice");
    globalThis.addEventListener(
      "UpdateSpotRemindersField",
      updateSpotRemindersField,
    );
  }, [setMode, setStartTime, updateSpotRemindersField]);

  const backToSetup = useCallback(() => {
    setMode("setup");
    document.removeEventListener(
      "UpdateSpotRemindersField",
      updateSpotRemindersField,
    );
  }, [updateSpotRemindersField]);

  useEffect(() => {
    // get initial sessions value from query param
    const urlParams = new URLSearchParams(window.location.search);
    const initialSessions = parseInt(urlParams.get("numSessions") ?? "", 10);
    if (!isNaN(initialSessions) && typeof initialSessions === "number") {
      setNumSessions(initialSessions);
    }
    if (initialspots) {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
      const initSpots: BasicSpot[] = JSON.parse(initialspots);
      if (!(initSpots instanceof Array) || initSpots.length === 0) {
        console.error("Invalid initial spots");
        return;
      }
      setSpots(initSpots);
    }
    const skipSetup = !!urlParams.get("skipSetup");
    if (skipSetup) {
      setShowPrepare(true);
      startPracticing();
    }
  }, [initialspots, setNumSessions, startPracticing]);

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
  const intervalRef = useRef<ReturnType<typeof setInterval>>();

  const openDialog = useCallback(() => {
    if (dialogRef.current) {
      globalThis.handleShowModal();
      dialogRef.current.showModal();
      setTimeElapsed(0);
      intervalRef.current = setInterval(() => {
        setTimeElapsed((timeElapsed) => timeElapsed + 1);
      }, 10);
    }
  }, [setTimeElapsed]);

  const closeDialog = useCallback(() => {
    globalThis.handleCloseModal();
    if (dialogRef.current) {
      clearInterval(intervalRef.current);
      if (dialogRef.current) {
        dialogRef.current.classList.add("close");
        setTimeout(() => {
          if (dialogRef.current) {
            dialogRef.current.close();
            dialogRef.current.classList.remove("close");
          }
        }, 150);
      }
    }
  }, []);

  useEffect(() => {
    if (show) {
      openDialog();
      return () => closeDialog();
    }
  }, [closeDialog, openDialog, show]);

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
          />
        </VioletButton>
      </div>
    </dialog>
  );
}

export function SingleSetupForm({
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
    if (numSessionsRef.current) {
      setNumSessions(parseInt(numSessionsRef.current.value, 10));
    }
    if (spots.length === 0) {
      if (!numSessionsRef.current) {
        setNumSessions(1);
      } else {
        setNumSessions(parseInt(numSessionsRef.current.value, 10));
      }
      globalThis.dispatchEvent(
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
  }, [spots.length, submit, setNumSessions]);

  const increaseSessions = useCallback(() => {
    if (!numSessionsRef.current) return;
    const curr = parseInt(numSessionsRef.current.value ?? "", 10);
    if (isNaN(curr) || curr < 1) {
      numSessionsRef.current.value = "1";
    }
    numSessionsRef.current.value = (curr + 1).toString();
  }, []);

  const decreaseSessions = useCallback(() => {
    if (!numSessionsRef.current) return;
    const curr = parseInt(numSessionsRef.current?.value, 10);
    if (isNaN(curr) || curr < 1) {
      numSessionsRef.current.value = "1";
    }
    numSessionsRef.current.value = (curr - 1).toString();
  }, []);

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
        <div className="flex-shrink-0 flex-grow" />
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
              className="basic-field w-full xs:w-20"
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
          <button
            type="button"
            onClick={handleSubmit}
            className="action-button violet focusable h-20 px-8 text-3xl"
          >
            Start Practicing
            <span className="icon-[iconamoon--player-play-thin] size-8" />
          </button>
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
  id: string;
};

function findNeglectedSpot(
  spots: SpotNeglectInfo[],
  eligibleSpotIds: string[],
): [boolean, string] {
  let maxReps = -Infinity;
  let minReps = Infinity;
  let minId = "";
  for (const spot of spots) {
    if (!eligibleSpotIds.includes(spot.id) || !spot.id) {
      continue;
    }
    if (spot.reps > maxReps) {
      maxReps = spot.reps;
    }
    if (spot.reps < minReps) {
      minReps = spot.reps;
      minId = spot.id;
    }
  }

  if (maxReps - minReps > 2) {
    return [true, minId];
  }
  return [false, ""];
}

// TODO: This should probably be a few components at this point

export function SinglePractice({
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
  const [currentSpot, setCurrentSpot] = useState<BasicSpot | null>(null);
  const [neglectInfo, setNeglectInfo] = useState<SpotNeglectInfo[]>([]);
  const [eligibleSpots, setEligibleSpots] = useState<BasicSpot[]>(spots);
  const [practiceSummary, setPracticeSummary] = useState<
    Map<string, { excellent: number; fine: number; poor: number }>
  >(new Map());
  // This counter ensures that the animation runs, even if the same spot is generated twice in a row.
  const [counter, setCounter] = useState(0);
  const [lastTwoSpots, setLastTwoSpots] = useState<string[]>([]);

  const [sessionsCompleted, setSessionsCompleted] = useState(0);
  const [sessionStarted, setSessionStarted] = useState(dayjs());
  const [hasShownResume, setHasShownResume] = useState(false);
  const [canContinue, setCanContinue] = useState(false);

  const resumeRef = useRef<HTMLDialogElement>(null);
  const topRef = useRef<HTMLDivElement>(null);
  const breakDialogRef = useRef<HTMLDialogElement>(null);
  const spotIdsHash = useMemo(() => {
    const ids = spots
      .map((spot) => spot.id)
      .sort()
      .join("");
    return hash(ids);
  }, [spots]);

  const saveToStorage = useCallback(
    (key: string, value: string) => {
      localStorage.setItem(`${pieceid}.${spotIdsHash}.${key}`, value);
      localStorage.setItem(
        `${pieceid}.${spotIdsHash}.savedAt`,
        Date.now().toString(),
      );
    },
    [pieceid, spotIdsHash],
  );

  const loadFromStorage = useCallback(
    (key: string) => {
      const savedAt = localStorage.getItem(`${pieceid}.${spotIdsHash}.savedAt`);
      if (
        !savedAt ||
        dayjs(parseInt(savedAt, 10)).isBefore(dayjs().subtract(1, "day"))
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
  }, []);

  const handleDone = useCallback(() => {
    const finalSummary: PracticeSummaryItem[] = [];
    for (const spot of spots) {
      if (!spot.id) {
        continue;
      }
      const results = practiceSummary.get(spot.id) ?? {
        excellent: 0,
        fine: 0,
        poor: 0,
      };
      let day = 0;
      if (spot.stageStarted) {
        const stageStarted = dayjs.unix(spot.stageStarted).tz(dayjs.tz.guess());
        const now = dayjs().tz(dayjs.tz.guess());
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
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.eligibleSpotIds`);
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.sessionsCompleted`);
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.savedAt`);
    }
    finish(finalSummary);
  }, [practiceSummary, finish, spots, pieceid, spotIdsHash]);

  const goToNextSpot = useCallback(
    (nextEligibleSpots: BasicSpot[]) => {
      setCounter((curr) => curr + 1);
      if (nextEligibleSpots.length === 0) {
        globalThis.dispatchEvent(
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

      const eligibleSpotIds: string[] = [];
      for (const spot of nextEligibleSpots) {
        if (!spot.id) {
          continue;
        }
        if (!lastTwoSpots.includes(spot.id)) {
          eligibleSpotIds.push(spot.id);
        }
      }
      const [hasNeglectedSpot, neglectedSpotId] = findNeglectedSpot(
        neglectInfo,
        eligibleSpotIds,
      );
      if (hasNeglectedSpot) {
        const nextSpot = nextEligibleSpots.find(
          (spot) => spot.id === neglectedSpotId,
        );
        if (nextSpot) {
          setCurrentSpot(nextSpot);
          setLastTwoSpots([neglectedSpotId, lastTwoSpots[0]]);
          return;
        }
        console.error("invalid neglected spot id");
      }
      const nextSpotId =
        eligibleSpotIds[Math.floor(Math.random() * eligibleSpotIds.length)];
      const nextSpot = nextEligibleSpots.find((spot) => spot.id === nextSpotId);
      if (!nextSpot) {
        console.error("invalid next spot id");
        return;
      }
      setCurrentSpot(nextSpot);
      if (nextSpotId) {
        setLastTwoSpots([nextSpotId, lastTwoSpots[0]]);
      }
    },
    [neglectInfo, handleDone, lastTwoSpots],
  );

  const handleResume = useCallback(() => {
    if (hasShownResume) {
      return;
    }
    setHasShownResume(true);
    const summary = loadFromStorage("practiceSummary");
    if (summary) {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
      setPracticeSummary(
        new Map(
          JSON.parse(summary) as [
            string,
            { excellent: number; fine: number; poor: number },
          ][],
        ),
      );
    } else {
      return;
    }
    const sessionsCompleted = loadFromStorage("sessionsCompleted");
    if (sessionsCompleted) {
      setSessionsCompleted(parseInt(sessionsCompleted, 10));
    }
    const esi = loadFromStorage("eligibleSpotIds");
    if (esi) {
      const eligibleSpotIds = JSON.parse(esi) as string[];
      const nextEligibleSpots = eligibleSpots.filter((spot) => {
        if (!spot.id) {
          return false;
        }
        return eligibleSpotIds.includes(spot.id);
      });
      if (nextEligibleSpots.length === 0) {
        goToNextSpot(eligibleSpots);
        return;
      }
      setEligibleSpots(nextEligibleSpots);
      goToNextSpot(nextEligibleSpots);
    } else {
      goToNextSpot(eligibleSpots);
    }
  }, [
    eligibleSpots,
    hasShownResume,
    loadFromStorage,
    goToNextSpot,
    setEligibleSpots,
  ]);

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
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.eligibleSpotIds`);
      localStorage.removeItem(`${pieceid}.${spotIdsHash}.savedAt`);
    }
  }, [
    spotIdsHash,
    pieceid,
    hasShownResume,
    loadFromStorage,
    handleResume,
    spots,
    setNeglectInfo,
    goToNextSpot,
    eligibleSpots,
  ]);

  useEffect(() => {
    if (!currentSpot) {
      const nextSpotIdx = Math.floor(Math.random() * eligibleSpots.length);
      const nextSpot = eligibleSpots[nextSpotIdx];
      setCurrentSpot(nextSpot);
    }
    if (neglectInfo.length === 0) {
      const spotNeglect: SpotNeglectInfo[] = [];
      for (const spot of spots) {
        if (!spot.id) {
          continue;
        }
        spotNeglect.push({
          id: spot.id,
          reps: 0,
        });
      }
      setNeglectInfo(spotNeglect);
    }
  }, [currentSpot, eligibleSpots, neglectInfo.length, spots]);

  const addSpotRep = useCallback(
    (id: string | undefined, quality: "excellent" | "fine" | "poor") => {
      const currentSpotIdx = neglectInfo.findIndex((spot) => spot.id === id);
      if (currentSpotIdx > -1) {
        const nextNeglectInfo = [...neglectInfo];
        nextNeglectInfo[currentSpotIdx].reps++;
        setNeglectInfo(nextNeglectInfo);
      }
      if (!id || !quality) {
        return;
      }
      const summary = practiceSummary.get(id) ?? {
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
    [practiceSummary, pieceid, saveToStorage, neglectInfo],
  );

  const evictSpot = useCallback(
    (spotId: string) => {
      // going to need a copy of this because the it won't be updated by setstate until after the function finishes
      const nextEligibleSpots = eligibleSpots.filter(
        (spot) => spot.id !== spotId,
      );
      if (pieceid) {
        saveToStorage(
          "eligibleSpotIds",
          JSON.stringify(nextEligibleSpots.map((spot) => spot.id)),
        );
      }
      setEligibleSpots(nextEligibleSpots);
      return nextEligibleSpots;
    },
    [eligibleSpots, pieceid, saveToStorage, setEligibleSpots],
  );

  const takeABreak = useCallback(() => {
    if (breakDialogRef.current) {
      breakDialogRef.current.showModal();
      globalThis.handleShowModal();
      setCanContinue(false);
      setTimeout(() => {
        setCanContinue(true);
      }, 30000);
      // }, 1000);
    }
  }, [setCanContinue]);

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

  const handleExcellent = useCallback(() => {
    const currentSpotId = currentSpot?.id;
    if (!currentSpotId) {
      return;
    }
    const summary = addSpotRep(currentSpotId, "excellent");
    if (!summary) {
      return;
    }
    let nextEligibleSpots = eligibleSpots;
    if (summary.excellent - summary.poor > 4) {
      nextEligibleSpots = evictSpot(currentSpotId);
    }
    maybeTakeABreak();
    goToNextSpot(nextEligibleSpots);
  }, [
    currentSpot?.id,
    addSpotRep,
    eligibleSpots,
    maybeTakeABreak,
    goToNextSpot,
    evictSpot,
  ]);

  const handleFine = useCallback(() => {
    const currentSpotId = currentSpot?.id;
    if (!currentSpotId) {
      return;
    }
    addSpotRep(currentSpotId, "fine");
    maybeTakeABreak();
    goToNextSpot(eligibleSpots);
  }, [
    currentSpot?.id,
    addSpotRep,
    maybeTakeABreak,
    goToNextSpot,
    eligibleSpots,
  ]);

  const handlePoor = useCallback(() => {
    const currentSpotId = currentSpot?.id;
    if (!currentSpotId) {
      return;
    }
    const summary = addSpotRep(currentSpotId, "poor");
    if (!summary) {
      return;
    }
    let nextEligibleSpots = eligibleSpots;
    if (summary.poor > 2) {
      nextEligibleSpots = evictSpot(currentSpotId);
    }
    maybeTakeABreak();
    goToNextSpot(nextEligibleSpots);
  }, [
    addSpotRep,
    currentSpot?.id,
    eligibleSpots,
    evictSpot,
    maybeTakeABreak,
    goToNextSpot,
  ]);

  const startSession = useCallback(() => {
    setSessionStarted(dayjs());
  }, [setSessionStarted]);

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
              currentSpot ? (
                <PracticeSpotDisplay spot={currentSpot} pieceid={pieceid} />
              ) : (
                <p className="text-center">Something went wrong</p>
              )
            }
            id={`${currentSpot?.id}-${counter}`}
          />
        </div>
        <div className="flex w-full flex-col justify-center gap-4 px-4 pt-8 xs:flex-row-reverse xs:px-0">
          <BigHappyButton type="button" onClick={handleExcellent}>
            <span
              className="icon-[iconamoon--like-thin] -ml-1 -mt-1 size-8"
              aria-hidden="true"
            />
            Excellent
          </BigHappyButton>
          <BigSkyButton type="button" onClick={handleFine}>
            <span
              className="icon-[iconamoon--sign-minus-thin] -ml-1 size-8"
              aria-hidden="true"
            />
            Fine
          </BigSkyButton>
          <BigAngryButton type="button" onClick={handlePoor}>
            <span
              className="icon-[iconamoon--dislike-thin] -mb-1 -ml-1 size-8"
              aria-hidden="true"
            />
            Poor
          </BigAngryButton>
        </div>
        <div className="flex justify-center gap-4 pb-12 pt-8">
          <BasicButton onClick={setup}>
            <span
              className="icon-[iconamoon--settings-thin] -ml-1 size-5"
              aria-hidden="true"
            />{" "}
            Back to setup
          </BasicButton>
          <WarningButton grow onClick={handleDone}>
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
