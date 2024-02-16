import { AnimatePresence, motion } from "framer-motion";
import { ScaleCrossFadeContent } from "../ui/transitions";
import { RepeatPrepareText } from "./repeat-prepare-text";
import { type BasicSpot } from "../validators";
import {
  BackToPiece,
  BackToPlan,
  HappyLink,
  Link,
  WarningLink,
} from "../ui/links";
import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { PracticeSpotDisplay } from "./practice-spot-display";
import { NextPlanItem } from "../ui/plan-components";
import { cn } from "../common";
import * as Switch from "@radix-ui/react-switch";

type RepeatMode = "prepare" | "practice" | "break_success" | "break_fail";

// TODO: add event listener to update spot
export function Repeat({
  initialspot,
  pieceid,
  csrf,
  piecetitle = "",
  planid = "",
  kidmode = true,
}: {
  initialspot?: string;
  pieceid?: string;
  csrf?: string;
  piecetitle?: string;
  planid?: string;
  kidmode?: boolean;
}) {
  const [mode, setMode] = useState<RepeatMode>("prepare");
  const [startTime, setStartTime] = useState<number>(0);
  const [spot, setSpot] = useState<BasicSpot | null>(null);
  const [kidMode, setKidMode] = useState<boolean>(kidmode);

  useEffect(() => {
    if (initialspot) {
      setSpot(JSON.parse(initialspot) as BasicSpot);
    }
    const savedKidMode = localStorage.getItem("kidmode");
    if (savedKidMode) {
      setKidMode(savedKidMode === "true");
    }
  }, [initialspot]);

  const startPracticing = useCallback(() => {
    setStartTime(Date.now());
    setMode("practice");
  }, [setMode, setStartTime]);

  const setModePrepare = useCallback(() => {
    setMode("prepare");
  }, [setMode]);

  const setModeBreakSuccess = useCallback(() => {
    setMode("break_success");
    if (
      spot?.stage !== "repeat" &&
      spot?.stage !== "extra_repeat" &&
      pieceid &&
      csrf &&
      startTime &&
      spot?.id
    ) {
      const durationMinutes = Math.ceil(
        (new Date().getTime() - startTime) / 1000 / 60,
      );
      globalThis.dispatchEvent(
        new CustomEvent("FinishedRepeatPracticing", {
          detail: {
            success: true,
            durationMinutes,
            csrf,
            toStage: "",
            endpoint: `/library/pieces/${pieceid}/spots/${spot.id}/practice/repeat`,
          },
        }),
      );
    }
  }, [setMode, pieceid, csrf, startTime, spot]);

  const updateKidMode = useCallback(
    (nextMode: boolean) => {
      localStorage.setItem("kidmode", nextMode ? "true" : "false");
      setKidMode(nextMode);
    },
    [setKidMode],
  );

  const promoteSpot = useCallback(
    (toStage: string) => {
      if (pieceid && csrf && startTime && spot?.id) {
        const durationMinutes = Math.ceil(
          (new Date().getTime() - startTime) / 1000 / 60,
        );
        globalThis.dispatchEvent(
          new CustomEvent("FinishedRepeatPracticing", {
            detail: {
              success: true,
              durationMinutes,
              csrf,
              toStage,
              endpoint: `/library/pieces/${pieceid}/spots/${spot.id}/practice/repeat`,
            },
          }),
        );
      }
    },
    [pieceid, csrf, startTime, spot],
  );

  const setModeBreakFail = useCallback(() => {
    if (pieceid && csrf && startTime && spot?.id) {
      const durationMinutes = Math.ceil(
        (new Date().getTime() - startTime) / 1000 / 60,
      );
      globalThis.dispatchEvent(
        new CustomEvent("FinishedRepeatPracticing", {
          detail: {
            success: false,
            durationMinutes,
            csrf,
            toStage: "",
            endpoint: `/library/pieces/${pieceid}/spots/${spot.id}/practice/repeat`,
          },
        }),
      );
    }
    setMode("break_fail");
  }, [setMode, pieceid, csrf, startTime, spot]);

  const updateSpot = useCallback((spot: BasicSpot) => {
    setSpot(spot);
  }, []);

  return (
    <div className="relative left-0 top-0 w-full sm:mx-auto sm:max-w-6xl">
      <ScaleCrossFadeContent
        component={
          {
            prepare: (
              <RepeatPrepare
                startPracticing={startPracticing}
                spot={spot}
                pieceid={pieceid}
                piecetitle={piecetitle}
                kidMode={kidMode}
                setKidMode={updateKidMode}
                csrf={csrf}
                updateSpot={updateSpot}
              />
            ),
            practice: (
              <RepeatPractice
                startTime={startTime}
                onSuccess={setModeBreakSuccess}
                onFail={setModeBreakFail}
                spot={spot}
                pieceid={pieceid}
                piecetitle={piecetitle}
                kidMode={kidMode}
                csrf={csrf}
                updateSpot={updateSpot}
              />
            ),
            break_success: (
              <RepeatBreakSuccess
                restart={setModePrepare}
                pieceid={pieceid}
                canPromote={
                  spot?.stage === "repeat" || spot?.stage === "extra_repeat"
                }
                promoteSpot={promoteSpot}
                planid={planid}
                kidMode={kidMode}
                csrf={csrf}
              />
            ),
            break_fail: (
              <RepeatBreakFail
                restart={setModePrepare}
                pieceid={pieceid}
                planid={planid}
                csrf={csrf}
              />
            ),
          }[mode]
        }
        id={mode}
      />
    </div>
  );
}

function RepeatPrepare({
  startPracticing,
  spot,
  pieceid,
  piecetitle = "",
  kidMode,
  setKidMode,
  csrf = "",
  updateSpot,
}: {
  startPracticing: () => void;
  spot?: BasicSpot | null;
  pieceid?: string;
  piecetitle?: string;
  kidMode: boolean;
  setKidMode: (kidMode: boolean) => void;
  csrf?: string;
  updateSpot: (spot: BasicSpot) => void;
}) {
  return (
    <div className="flex w-full flex-col" id="repeat-prepare-wrapper">
      <RepeatPracticeTitle stage={spot?.stage} />
      <div className="py-2">
        {spot && (
          <PracticeSpotDisplay
            spot={spot}
            pieceid={pieceid}
            piecetitle={piecetitle}
            csrf={csrf}
            updateSpot={updateSpot}
          />
        )}
      </div>

      <div className="col-span-full flex w-full items-center justify-center py-8">
        <button
          type="button"
          onClick={startPracticing}
          className="action-button violet focusable h-20 px-8 text-3xl"
        >
          Go Practice
          <span className="icon-[iconamoon--player-play-thin] size-8" />
        </button>
      </div>
      <div className="flex justify-center">
        <div className="mt-2 flex flex-col gap-2 rounded-xl bg-neutral-700/5 p-4">
          <label
            htmlFor="visual-mode"
            className="w-full text-center text-2xl font-bold"
          >
            Visual Mode
          </label>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setKidMode(false)}
              className={cn(
                "flex items-center gap-1 border-b-2 pb-2 text-xl font-semibold",
                kidMode
                  ? "border-transparent text-neutral-800"
                  : "border-green-500 text-green-600",
              )}
            >
              Simple
              <span
                className={cn(
                  "inline-flex size-10 select-none items-center justify-center rounded-xl border p-2 shadow",
                  kidMode
                    ? "border-neutral-500 bg-neutral-300 text-neutral-500 shadow-neutral-900/50"
                    : "border border-green-600 bg-green-300 text-green-600 shadow-green-900/50",
                )}
              >
                <span className="icon-[iconamoon--check-bold] size-6 sm:size-8 lg:size-12" />
              </span>
            </button>
            <div className="border-b-2 border-transparent pb-2">
              <Switch.Root
                className={cn(
                  "focusable relative h-[25px] w-[42px] cursor-default rounded-full shadow shadow-neutral-900/20 outline-none",
                  kidMode ? "bg-yellow-500" : "bg-green-600",
                )}
                id="visual-mode"
                name="visual-mode"
                style={{ "-webkit-tap-highlight-color": "rgba(0, 0, 0, 0)" }}
                onCheckedChange={() => setKidMode(!kidMode)}
                checked={kidMode}
              >
                <Switch.Thumb className="shadow-blackA4 block h-[21px] w-[21px] translate-x-0.5 rounded-full bg-white shadow-[0_2px_2px] transition-transform duration-100 will-change-transform data-[state=checked]:translate-x-[19px]" />
              </Switch.Root>
            </div>
            <button
              onClick={() => setKidMode(true)}
              className={cn(
                "flex items-center gap-1 border-b-2 pb-2 text-xl font-semibold",
                kidMode
                  ? "border-yellow-500 text-yellow-600"
                  : "border-transparent text-neutral-800",
              )}
            >
              <span
                className={cn(
                  "inline-flex size-10 select-none items-center justify-center rounded-xl border p-2 shadow",
                  kidMode
                    ? "border-yellow-500 bg-yellow-300 text-yellow-500 shadow-yellow-900/50 "
                    : "border-neutral-500 bg-neutral-300 text-neutral-500 shadow-neutral-900/20",
                )}
              >
                <span className="icon-[iconamoon--star-duotone] size-6 sm:size-8 lg:size-12" />
              </span>
              Fun
            </button>
          </div>
        </div>
      </div>
      <RepeatPrepareText open={!spot} />
    </div>
  );
}

function RepeatPracticeTitle(props: { stage?: string | null }) {
  switch (props.stage) {
    case "repeat":
      return (
        <>
          <h2 className="py-1 text-left text-2xl font-bold">New Spot</h2>
          <p className="text-left text-base">
            Make a plan to play this spot correctly, then start practicing.
          </p>
        </>
      );
    case "extra_repeat":
      return (
        <>
          <h2 className="py-1 text-left text-2xl font-bold">
            Extra Repeat Practicing
          </h2>
          <p className="text-left text-base">
            Try to remember what made this spot successful last time so you can
            recreate that
          </p>
        </>
      );
    default:
      return (
        <>
          <h2 className="py-1 text-left text-2xl font-bold">
            Repeat Practicing
          </h2>
          <p className="text-left text-base">
            Repeat practicing is an important part of learning, but you need to
            do it carefully!
          </p>
        </>
      );
  }
}

function RepeatPractice({
  onSuccess,
  onFail,
  startTime,
  spot,
  pieceid,
  piecetitle = "",
  kidMode = false,
  csrf = "",
  updateSpot,
}: {
  onSuccess: () => void;
  onFail: () => void;
  startTime: number;
  spot?: BasicSpot | null;
  pieceid?: string;
  piecetitle?: string;
  kidMode?: boolean;
  csrf?: string;
  updateSpot: (spot: BasicSpot) => void;
}) {
  const [numCompleted, setCompleted] = useState(0);
  const [waitedLongEnough, setWaitedLongEnough] = useState(true);
  const topRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (topRef.current) {
      window.scrollTo(0, topRef.current.offsetTop - 160);
    }
  }, []);

  const succeed = useCallback(() => {
    if (numCompleted === 4) {
      setCompleted((curr) => curr + 1);
      setTimeout(onSuccess, 300);
    } else {
      setCompleted((curr) => curr + 1);
      setWaitedLongEnough(false);
      setTimeout(() => {
        setWaitedLongEnough(true);
      }, 1000);
    }
  }, [numCompleted, onSuccess, setWaitedLongEnough]);

  const fail = useCallback(() => {
    setCompleted(0);
    setWaitedLongEnough(false);
    setTimeout(() => {
      setWaitedLongEnough(true);
    }, 1500);
    // if it's been more than five minutes + 30 seconds buffer, take a break
    if (Date.now() - startTime > 5 * 60 * 1000 + 30 * 1000) {
      onFail();
    }
  }, [setCompleted, onFail, startTime]);

  return (
    <>
      <div
        className="flex w-full flex-col sm:mx-auto sm:max-w-3xl"
        ref={topRef}
      >
        <div className="flex w-full flex-col items-center sm:mx-auto sm:max-w-xl">
          <h2 className="text-center text-2xl font-semibold">Completions</h2>
          <ul className="my-4 grid grid-cols-5 gap-4">
            <PracticeListItem
              num={1}
              completed={numCompleted >= 1}
              kidMode={kidMode}
            />
            <PracticeListItem
              num={2}
              completed={numCompleted >= 2}
              kidMode={kidMode}
            />
            <PracticeListItem
              num={3}
              completed={numCompleted >= 3}
              kidMode={kidMode}
            />
            <PracticeListItem
              num={4}
              completed={numCompleted >= 4}
              kidMode={kidMode}
            />
            <PracticeListItem
              num={5}
              completed={numCompleted >= 5}
              kidMode={kidMode}
            />
          </ul>
        </div>
        <div className="flex w-full flex-col items-center pt-4 sm:mx-auto sm:max-w-xl">
          <h2 className="text-center text-3xl font-semibold">How did it go?</h2>
          <div className="flex w-full flex-col justify-center gap-4 px-4 pt-8 xs:flex-row-reverse xs:px-0">
            <button
              disabled={!waitedLongEnough}
              onClick={succeed}
              className="action-button focusable green h-16 px-6 py-4 text-2xl"
            >
              <span className="icon-[iconamoon--like-thin] -ml-1 -mt-1 size-8" />
              <span>Correct</span>
            </button>
            <button
              disabled={!waitedLongEnough}
              onClick={fail}
              className="action-button focusable red h-16 px-6 py-4 text-2xl"
            >
              <span className="icon-[iconamoon--dislike-thin] -mb-1 -ml-1 size-8" />
              <span>Mistake</span>
            </button>
          </div>
        </div>
      </div>
      {spot && (
        <div className="pt-4 md:px-8 md:pt-8">
          <PracticeSpotDisplay
            spot={spot}
            pieceid={pieceid}
            piecetitle={piecetitle}
            updateSpot={updateSpot}
            csrf={csrf}
          />
        </div>
      )}
      <div className="flex w-full flex-col py-12 sm:mx-auto sm:max-w-3xl">
        <div className="mx-auto flex w-full max-w-lg flex-wrap items-center justify-center">
          <button
            onClick={onFail}
            className="action-button red focusable"
            type="button"
          >
            <span>Move On</span>
            <span className="icon-[iconamoon--arrow-top-right-1-thin] -ml-1 size-6" />
          </button>
        </div>
      </div>
    </>
  );
}

const variants = {
  initial: {
    scale: 1.2,
  },
  animate: {
    scale: 1,
    transition: { bounce: 0, duration: 0.1, ease: "easeIn" },
  },
  exit: {
    scale: 1.2,
    transition: { bounce: 0, duration: 0.1, ease: "easeOut" },
  },
};

type RandomConfettiPos = "tl" | "tr" | "bl" | "br" | "t" | "b" | "l" | "r";
const confettiColors = [
  "#fbbf24",
  "#fbbf24",
  "#fbbf24",
  "#fbbf24",
  "#f87171",
  "#fbbf24",
  "#4ade80",
  "#fbbf24",
  "#60a5fa",
  "#a78bfa",
  "#f472b6",
] as const;

function randomConfettiStyle(
  pos: RandomConfettiPos,
  big: boolean,
  delay: number,
) {
  let top: number | string | undefined;
  let right: number | string | undefined;
  let bottom: number | string | undefined;
  let left: number | string | undefined;

  let xinit = Math.floor(Math.random() * 40);
  let yinit = Math.floor(Math.random() * 40);
  let cy = Math.floor(Math.random() * 40) + 10;
  let cx = Math.floor(Math.random() * 40) + 10;
  const color =
    confettiColors[Math.floor(Math.random() * confettiColors.length)];

  let duration = Math.floor(Math.random() * 500) + 500;

  let rot = Math.floor(Math.random() * 360) + 180;
  if (Math.random() > 0.5) {
    rot *= -1;
  }

  if (big) {
    cy *= 5;
    cx *= 5;
    xinit *= 6;
    yinit *= 6;
    duration *= 1.5;
  }

  switch (pos) {
    case "tl":
      cy = -cy;
      cx = -cx;
      top = 0;
      left = 0;
      break;
    case "tr":
      cy = -cy;
      top = 0;
      right = 0;
      break;
    case "bl":
      cx = -cx;
      bottom = 0;
      left = 0;
      break;
    case "br":
      bottom = 0;
      right = 0;
      break;
    case "t":
      top = 0;
      cy = -cy;
      left = "50%";
      break;
    case "b":
      bottom = 0;
      left = "50%";
      break;
    case "l":
      left = 0;
      top = "50%";
      cx = -cx;
      break;
    case "r":
      top = "50%";
      right = 0;
      break;
    default:
      break;
  }

  return {
    "--cy": `${cy}px`,
    "--cx": `${cx}px`,
    "--crot": `${rot}deg`,
    animationName: "confetti",
    animationDuration: `${duration}ms`,
    animationDirection: "normal",
    animationFillMode: "forwards",
    animationTimingFunction: "ease-in",
    animationDelay: `${delay}ms`,
    "--cyinit": `${yinit}px`,
    "--cxinit": `${xinit}px`,
    position: "absolute",
    top,
    right,
    left,
    bottom,
    color,
    textShadow: "1px 1px 2px black",
  };
}

const randomConfettiPositions: RandomConfettiPos[] = [
  "tl",
  "tr",
  "br",
  "bl",
  "t",
  "r",
  "b",
  "l",
];

function RandomStarConfetti(props: {
  id: string;
  big?: boolean;
  delay?: boolean;
}) {
  const [delay, setDelay] = useState(0);
  const [visibility, setVisibility] = useState("hidden");

  useEffect(() => {
    if (props.delay) {
      const d = Math.floor(Math.random() * 1000);
      setDelay(d);
      setTimeout(() => {
        setVisibility("visible");
      }, d + 20);
    } else {
      setDelay(0);
      setVisibility("visible");
    }
  }, [props.delay]);

  return (
    <>
      {[...randomConfettiPositions, ...randomConfettiPositions].map(
        (pos, i) => (
          <div
            key={`${pos}-${i}-${props.id}`}
            id={`${pos}-${i}-${props.id}`}
            className={cn(
              "icon-[iconamoon--star-duotone] ",
              props.big ? "size-6" : "size-2 lg:size-3",
            )}
            style={{
              ...randomConfettiStyle(pos, props.big ?? false, delay),
              visibility,
            }}
            aria-hidden="true"
          />
        ),
      )}
    </>
  );
}

function getRandomConfettiPositions() {
  const len = Math.floor(Math.random() * 10) + 10;
  const arr: [number, number][] = Array<[number, number]>(len).fill([0, 0]);
  arr[0] = [10, 10];
  arr[0] = [90, 10];
  for (let i = 2; i < len; ++i) {
    // these will be between 15 and 85, which will be used as top and left positions in view units
    arr[i] = [
      Math.floor(Math.random() * 75) + 10,
      Math.floor(Math.random() * 60) + 10,
    ];
  }
  return arr;
}

function PracticeListItem({
  num,
  completed,
  kidMode = false,
}: {
  num: number;
  completed: boolean;
  kidMode?: boolean;
}) {
  return (
    <AnimatePresence initial={false} mode="wait">
      {completed ? (
        <motion.li
          // @ts-expect-error It thinks it can't take a classname but it can
          className={cn(
            "relative flex size-10 select-none items-center justify-center rounded-xl border shadow transition-all duration-100 sm:size-12 lg:size-16",
            kidMode
              ? "border-yellow-500 bg-yellow-300 text-yellow-500 shadow-yellow-700/50 "
              : "border-green-600 bg-green-300 text-green-600 shadow-green-900/50 ",
          )}
          key={`${num}-completed`}
          initial="initial"
          animate="animate"
          exit="exit"
          variants={variants}
        >
          {kidMode ? (
            <>
              <RandomStarConfetti id={`confetti-${num}`} />

              <span className="icon-[iconamoon--star-duotone] z-10 size-6 sm:size-8 lg:size-12" />
            </>
          ) : (
            <span className="icon-[iconamoon--check-fill] size-6 sm:size-8 lg:size-12" />
          )}
          <span className="sr-only">Checked</span>
        </motion.li>
      ) : (
        <motion.li
          // @ts-expect-error It thinks it can't take a classname but it can
          className="flex size-10 select-none items-center justify-center rounded-xl border border-neutral-400 bg-neutral-200 text-neutral-400 opacity-80 transition-all duration-100 sm:size-12 lg:size-16"
          key={`${num}-incomplete`}
          initial="initial"
          animate="animate"
          exit="exit"
          variants={variants}
        >
          <div className="m-0 p-0 text-2xl font-medium sm:text-3xl lg:text-4xl">
            {num}
          </div>
        </motion.li>
      )}
    </AnimatePresence>
  );
}

function RepeatBreakSuccess({
  restart,
  pieceid,
  canPromote,
  promoteSpot,
  planid = "",
  kidMode = false,
  csrf,
}: {
  restart: () => void;
  pieceid?: string;
  canPromote: boolean;
  promoteSpot: (toStage: string) => void;
  planid?: string;
  kidMode?: boolean;
  csrf?: string;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [showConfetti, setShowConfetti] = useState(false);

  const close = useCallback(() => {
    globalThis.handleCloseModal();
    if (dialogRef.current) {
      if (dialogRef.current) {
        dialogRef.current.classList.add("close");
        setTimeout(() => {
          if (dialogRef.current) {
            dialogRef.current.classList.remove("close");
            dialogRef.current.close();
            setShowConfetti(true);
          }
        }, 150);
      }
    }
  }, [dialogRef]);

  const handleRandom = useCallback(() => {
    promoteSpot("random");
    close();
  }, [close, promoteSpot]);
  const handleMoreRepeat = useCallback(() => {
    promoteSpot("extra_repeat");
    close();
  }, [close, promoteSpot]);

  useEffect(() => {
    if (dialogRef.current) {
      if (canPromote) {
        dialogRef.current.showModal();
        globalThis.handleShowModal();
      } else {
        setShowConfetti(true);
      }
    }
  }, [canPromote, setShowConfetti]);

  return (
    <>
      {kidMode && showConfetti ? (
        <>
          {getRandomConfettiPositions().map(([x, y], i) => (
            <div
              class="absolute"
              key={i}
              style={{ top: `${y}vh`, left: `${x}vw` }}
            >
              <RandomStarConfetti id={`confetti-${i}`} big delay={i !== 1} />
            </div>
          ))}
        </>
      ) : null}
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
            Promote Spot
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          Great job! You can now move this spot to random practicing, or keep it
          in repeat for a little longer.
        </div>
        <div className="mt-2 flex w-full flex-col gap-2 sm:flex-row-reverse sm:gap-2">
          <button
            onClick={handleRandom}
            className="action-button focusable green w-full flex-grow text-lg"
            type="button"
          >
            <span className="icon-[iconamoon--playlist-shuffle-thin] -ml-1 size-6" />
            Random
          </button>
          <button
            onClick={handleMoreRepeat}
            className="action-button focusable sky focusable w-full flex-grow text-lg"
          >
            <span className="icon-[iconamoon--playlist-repeat-list-thin] -ml-1 size-6" />
            More Repeat
          </button>
        </div>
      </dialog>
      <div className="flex w-full flex-col items-center sm:mx-auto sm:max-w-3xl">
        <h1 className="py-1 text-center text-2xl font-bold">You did it!</h1>
        <p className="text-center text-base">
          Great job completing your five times in a row!
        </p>
        <div className="my-8 flex w-full flex-col justify-center gap-4 sm:flex-row sm:gap-6">
          <RepeatFinishedActionButtons
            pieceid={pieceid}
            planid={planid}
            restart={restart}
            csrf={csrf}
          />
        </div>
        <div className="prose prose-neutral mt-8">
          <h3 className="text-left text-lg">What to do now?</h3>
          <p className="text-sm">Here are a few options for what to do next.</p>
          <ul>
            <li>
              Take a moment to reflect on what allowed you to do this
              successfully so you can recreate it in the future.
            </li>
            <li>Take a break to let your brain recover</li>
            <li>
              Play this spot again in a few minutes with the goal of playing it
              correctly on the first try.
            </li>
            <li>Add this spot to your random practicing.</li>
            <li>Repeat practice another spot.</li>
          </ul>
        </div>
      </div>
    </>
  );
}

function RepeatBreakFail({
  restart,
  pieceid,
  planid = "",
  csrf = "",
}: {
  restart: () => void;
  pieceid?: string;
  planid?: string;
  csrf?: string;
}) {
  return (
    <div className="flex w-full flex-col items-center sm:mx-auto sm:max-w-3xl">
      <h1 className="py-1 text-center text-2xl font-bold">Time for a Break</h1>
      <p className="text-center text-base">
        You must put limits on your practicing so you don’t accidentally
        reinforce mistakes
      </p>

      <div className="my-8 flex w-full flex-col justify-center gap-4 sm:flex-row sm:gap-6">
        <RepeatFinishedActionButtons
          csrf={csrf}
          pieceid={pieceid}
          planid={planid}
          restart={restart}
        />
      </div>
      <div className="prose prose-neutral mt-8">
        <h3 className="text-left text-lg">What to do now?</h3>
        <p className="text-sm">Here are a few options for what to do next.</p>
        <ul>
          <li>
            <strong className="font-semibold">Reflect</strong> on this spot
            <ul>
              <li>Is it a technical or mental problem holding you back?</li>
              <li>Could you slow down more and have more success?</li>
              <li>
                Are you taking long enough breaks between repetitions to
                reflect?
              </li>
              <li>Would this be better as two smaller spots?</li>
              <li>
                Is it worth returning to this today or should you wait until
                tomorrow?
              </li>
            </ul>
          </li>
          <li>
            <strong className="font-semibold">Take a break!</strong> It can be
            just a few minutes, or if you’re really frustrated, make it an hour.
          </li>
          <li>
            <strong className="font-semibold">Move on</strong> to another spot
            or piece.
          </li>
          <li>
            <strong className="font-semibold">Come back later</strong> when you
            have a plan to be more successful.{" "}
          </li>
        </ul>
      </div>
    </div>
  );
}

function RepeatFinishedActionButtons({
  pieceid,
  planid,
  restart,
  csrf,
}: {
  pieceid?: string;
  planid?: string;
  restart?: () => void;
  csrf?: string;
}) {
  if (planid) {
    return (
      <>
        {!!pieceid && <BackToPiece pieceid={pieceid} />}
        <BackToPlan planid={planid} />
        <NextPlanItem planid={planid} csrf={csrf} />
      </>
    );
  }
  if (pieceid) {
    return (
      <>
        <BackToPiece pieceid={pieceid} />
        <WarningLink href={`/library/pieces/${pieceid}/practice/random-single`}>
          <span
            className="icon-[iconamoon--playlist-shuffle-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Try Random Practicing
        </WarningLink>
        <HappyLink href={`/library/pieces/${pieceid}/practice/repeat`}>
          <span
            className="icon-[iconamoon--playlist-repeat-list-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Practice Another Spot
        </HappyLink>
      </>
    );
  }
  return (
    <>
      <Link className="focusable action-button sky" href="/library">
        <span
          className="icon-[custom--music-note-screen] -ml-1 size-5"
          aria-hidden="true"
        />
        Library
      </Link>
      <WarningLink href="/practice/random-single">
        <span
          className="icon-[iconamoon--playlist-shuffle-thin] -ml-1 size-5"
          aria-hidden="true"
        />
        Try Random Practicing
      </WarningLink>
      <button
        onClick={restart}
        type="button"
        className="focusable action-button green"
      >
        <span
          className="icon-[iconamoon--playlist-repeat-list-thin] -ml-1 size-5"
          aria-hidden="true"
        />
        Practice Another Spot
      </button>
    </>
  );
}
