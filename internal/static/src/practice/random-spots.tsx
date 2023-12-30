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
  AngryButton,
  BasicButton,
  BigAngryButton,
  BigHappyButton,
  BigSkyButton,
  GiantBasicButton,
  HappyButton,
  WarningButton,
} from "../ui/buttons";
import {
  BookmarkIcon,
  CheckCircleIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  HandRaisedIcon,
  HandThumbDownIcon,
  HandThumbUpIcon,
  StopCircleIcon,
  XCircleIcon,
} from "@heroicons/react/24/solid";
import Summary from "./summary";
import { PracticeSpotDisplay } from "./practice-spot-display";
import dayjs from "dayjs";
import { BackToPlan } from "../ui/links";

export function RandomSpots({
  initialspots,
  pieceid,
  csrf,
  planid,
}: {
  initialspots?: string;
  pieceid?: string;
  csrf?: string;
  planid?: string;
}) {
  const [spots, setSpots] = useState<BasicSpot[]>([]);
  const [summary, setSummary] = useState<PracticeSummaryItem[]>([]);
  const [mode, setMode] = useState<RandomMode>("setup");
  const [startTime, setStartTime] = useState<Date | null>(null);
  const [numSessions, setNumSessions] = useState(2);

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
      // get initial sessions value from query param
      const urlParams = new URLSearchParams(window.location.search);
      const initialSessions = parseInt(urlParams.get("numSessions"));
      if (initialSessions && typeof initialSessions === "number") {
        setNumSessions(initialSessions);
      }
    },
    [initialspots, setNumSessions],
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
          <div className="flex gap-2">
            <input
              id="num-sessions"
              className="focusable w-20 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20"
              type="number"
              min="1"
              defaultValue={`${numSessions}`}
              ref={numSessionsRef}
            />
          </div>
        </div>
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

function hash(s: string) {
  for (var i = 0, h = 0xdeadbeef; i < s.length; i++)
    h = Math.imul(h ^ s.charCodeAt(i), 2654435761);
  return (h ^ (h >>> 16)) >>> 0;
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
  const resumeRef = useRef<HTMLDialogElement>(null);
  const topRef = useRef<HTMLDivElement>(null);
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
  const breakDialogRef = useRef<HTMLDialogElement>(null);
  const [canContinue, setCanContinue] = useState(false);
  const interleaveSpotsRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (interleaveSpotsRef.current) {
      // @ts-ignore
      htmx.process(interleaveSpotsRef.current);
    }
  }, [interleaveSpotsRef.current]);

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
      closeResume();
    },
    [resumeRef, setSkipSpotIds, setPracticeSummary],
  );

  const closeResume = useCallback(
    function () {
      if (resumeRef.current) {
        resumeRef.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (resumeRef.current) {
              resumeRef.current.classList.remove("close");
              resumeRef.current.close();
            }
          });
        });
      }
    },
    [resumeRef.current],
  );

  const closeBreak = useCallback(
    function () {
      if (breakDialogRef.current) {
        breakDialogRef.current.classList.add("close");
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            if (breakDialogRef.current) {
              breakDialogRef.current.classList.remove("close");
              breakDialogRef.current.close();
            }
          });
        });
      }
    },
    [breakDialogRef.current],
  );

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
      if (pieceid) {
        saveToStorage(
          "practiceSummary",
          JSON.stringify(Array.from(practiceSummary.entries())),
        );
      }
      setPracticeSummary(practiceSummary);
      return summary;
    },
    [setPracticeSummary, practiceSummary, pieceid, spotIdsHash],
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
        localStorage.removeItem(`${pieceid}.${spotIdsHash}.savedAt`);
      }
      finish(finalSummary);
    },
    [practiceSummary, finish, spots, pieceid, spotIdsHash],
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
      if (pieceid) {
        saveToStorage("skipSpotIds", JSON.stringify(newSkipSpotIds));
      }
      setSkipSpotIds(newSkipSpotIds);
      return newSkipSpotIds;
    },
    [spots, skipSpotIds, pieceid, spotIdsHash],
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
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, nextSpot, evictSpot],
  );

  const handleFine = useCallback(
    function () {
      const currentSpotId = spots[currentSpotIdx]?.id;
      addSpotRep(currentSpotId, "fine");
      maybeTakeABreak();
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
      maybeTakeABreak();
      nextSpot(nextSkipSpotIds);
    },
    [spots, currentSpotIdx, skipSpotIds, addSpotRep, nextSpot, evictSpot],
  );

  const startSession = useCallback(
    function () {
      closeBreak();
      setSessionStarted(dayjs());
    },
    [setSessionStarted],
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

  const takeABreak = useCallback(
    function () {
      if (breakDialogRef.current) {
        breakDialogRef.current.showModal();
        setCanContinue(false);
        setTimeout(function () {
          setCanContinue(true);
        }, 60000);
        // }, 1000);
      }
    },
    [breakDialogRef.current, setCanContinue],
  );

  // TODO: handle pretty widths for buttons
  return (
    <div className="relative w-full" ref={topRef}>
      <dialog
        ref={breakDialogRef}
        aria-labelledby="resume-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="resume-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Time for a Break
          </h3>
        </header>
        <div className="mt-2 flex flex-col gap-2 text-left text-neutral-700">
          <p>
            It’s been five minutes, time for a short break. Let your mind and
            hands relax.
          </p>
          {!!planid && (
            <>
              <p>This is a great time to practice your interleave spots!</p>
              <details>
                <summary className="flex cursor-pointer items-center justify-between gap-1 rounded-xl bg-indigo-500/50 py-2 pl-4 pr-2 font-semibold text-indigo-800 transition duration-200 hover:bg-indigo-300/50">
                  <div className="flex items-center gap-2">
                    <BookmarkIcon className="-ml-1 size-5" />
                    Interleave Spots
                  </div>
                  <ChevronRightIcon className="summary-icon size-6 transition-transform" />
                </summary>
                <div
                  ref={interleaveSpotsRef}
                  hx-trigger="load"
                  hx-swap="innterHTML transition:true"
                  hx-get={`/library/plans/${planid}/interleave`}
                  className="w-full py-2"
                >
                  Loading Interleave Spots...
                </div>
              </details>
              <p>
                You can also go back to your practice plan and resume this
                later.
              </p>
            </>
          )}
        </div>
        {canContinue ? (
          <div className="flex w-full flex-col flex-wrap gap-2 sm:flex-row-reverse">
            <HappyButton
              grow
              onClick={startSession}
              className="text-lg"
              disabled={!canContinue}
            >
              <CheckCircleIcon className="-ml-1 size-5" />
              Continue
            </HappyButton>
            <WarningButton grow onClick={handleDone}>
              <StopCircleIcon className="-ml-1 size-5" />
              Finish
            </WarningButton>
            {!!planid && <BackToPlan planid={planid} grow />}
          </div>
        ) : (
          <div className="mt-1 flex w-full justify-center py-2 text-lg font-medium">
            Enjoy your break!
          </div>
        )}
      </dialog>

      <dialog
        ref={resumeRef}
        aria-labelledby="resume-title"
        className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
      >
        <header className="mt-2 text-center sm:text-left">
          <h3
            id="resume-title"
            className="text-2xl font-semibold leading-6 text-neutral-900"
          >
            Resume Practicing
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          It looks like your practicing was interrupted, would you like to
          resume it?
        </div>
        <div className="mt-2 flex w-full flex-row-reverse flex-wrap gap-2 sm:gap-2">
          <HappyButton grow onClick={handleResume} className="text-lg">
            <CheckCircleIcon className="-ml-1 size-5" />
            Resume
          </HappyButton>
          <AngryButton grow onClick={closeResume} className="text-lg">
            <XCircleIcon className="-ml-1 size-5" />
            Close
          </AngryButton>
        </div>
      </dialog>
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
          <BasicButton onClick={setup}>
            <ChevronLeftIcon className="-ml-1 size-5" /> Back to setup
          </BasicButton>
          <WarningButton grow onClick={handleDone}>
            <StopCircleIcon className="-ml-1 size-5" />
            Finish
          </WarningButton>
        </div>
      </div>
    </div>
  );
}
