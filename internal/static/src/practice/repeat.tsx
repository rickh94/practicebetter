import { AnimatePresence, motion } from "framer-motion";
import { ScaleCrossFadeContent } from "../ui/transitions";
import { RepeatPrepareText } from "./repeat-prepare-text";
import {
  BigAngryButton,
  BigHappyButton,
  GiantBasicButton,
  HappyButton,
  SkyButton,
  WarningButton,
} from "../ui/buttons";
import { BasicSpot } from "../validators";
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

type RepeatMode = "prepare" | "practice" | "break_success" | "break_fail";

// TODO: add event listener to update spot
export function Repeat({
  initialspot,
  pieceid,
  csrf,
  piecetitle = "",
  planid = "",
}: {
  initialspot?: string;
  pieceid?: string;
  csrf?: string;
  piecetitle?: string;
  planid?: string;
}) {
  const [mode, setMode] = useState<RepeatMode>("prepare");
  const [startTime, setStartTime] = useState<number>(0);
  const [spot, setSpot] = useState<BasicSpot | null>(null);

  useEffect(
    function () {
      if (initialspot) {
        setSpot(JSON.parse(initialspot));
      }
    },
    [initialspot],
  );

  const startPracticing = useCallback(
    function () {
      setStartTime(Date.now());
      setMode("practice");
    },
    [setMode, setStartTime],
  );

  const setModePrepare = useCallback(
    function () {
      setMode("prepare");
    },
    [setMode],
  );

  const setModeBreakSuccess = useCallback(
    function () {
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
        document.dispatchEvent(
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
    },
    [setMode, pieceid, csrf, startTime, spot, spot?.id],
  );

  const promoteSpot = useCallback(
    function (toStage: string) {
      if (pieceid && csrf && startTime && spot?.id) {
        const durationMinutes = Math.ceil(
          (new Date().getTime() - startTime) / 1000 / 60,
        );
        document.dispatchEvent(
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
    [pieceid, csrf, startTime, spot, spot?.id],
  );

  const setModeBreakFail = useCallback(
    function () {
      if (pieceid && csrf && startTime && spot?.id) {
        const durationMinutes = Math.ceil(
          (new Date().getTime() - startTime) / 1000 / 60,
        );
        document.dispatchEvent(
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
    },
    [setMode, pieceid, csrf, startTime, spot, spot?.id],
  );

  return (
    <div className="relative left-0 top-0 w-full sm:mx-auto sm:max-w-5xl">
      <ScaleCrossFadeContent
        component={
          {
            prepare: (
              <RepeatPrepare
                startPracticing={startPracticing}
                spot={spot}
                pieceid={pieceid}
                piecetitle={piecetitle}
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
              />
            ),
            break_fail: (
              <RepeatBreakFail
                restart={setModePrepare}
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

function RepeatPrepare({
  startPracticing,
  spot,
  pieceid,
  piecetitle = "",
}: {
  startPracticing: () => void;
  spot?: BasicSpot;
  pieceid?: string;
  piecetitle?: string;
}) {
  return (
    <div className="flex w-full flex-col" id="repeat-prepare-wrapper">
      <h2 className="py-1 text-left text-2xl font-bold">Repeat Practicing</h2>
      <p className="text-left text-base">
        Repeat practicing is an important part of learning, but you need to do
        it carefully!
      </p>
      <div className="py-2">
        {spot && (
          <PracticeSpotDisplay
            spot={spot}
            pieceid={pieceid}
            piecetitle={piecetitle}
          />
        )}
      </div>
      <RepeatPrepareText open={!spot} />
      <div className="col-span-full flex w-full items-center justify-center py-16">
        <GiantBasicButton type="button" onClick={startPracticing}>
          Start Practicing
        </GiantBasicButton>
      </div>
    </div>
  );
}

function RepeatPractice({
  onSuccess,
  onFail,
  startTime,
  spot,
  pieceid,
  piecetitle = "",
}: {
  onSuccess: () => void;
  onFail: () => void;
  startTime: number;
  spot?: BasicSpot;
  pieceid?: string;
  piecetitle?: string;
}) {
  const [numCompleted, setCompleted] = useState(0);
  const [waitedLongEnough, setWaitedLongEnough] = useState(true);
  const topRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (topRef.current) {
      window.scrollTo(0, topRef.current.offsetTop - 160);
    }
  }, [topRef.current]);

  const succeed = useCallback(
    function () {
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
    },
    [numCompleted, onSuccess, setWaitedLongEnough],
  );

  const fail = useCallback(
    function () {
      setCompleted(0);
      setWaitedLongEnough(false);
      setTimeout(() => {
        setWaitedLongEnough(true);
      }, 1500);
      // if it's been more than five minutes + 30 seconds buffer, take a break
      if (Date.now() - startTime > 5 * 60 * 1000 + 30 * 1000) {
        onFail();
      }
    },
    [setCompleted, onFail, startTime],
  );

  return (
    <>
      <div
        className="flex w-full flex-col sm:mx-auto sm:max-w-3xl"
        ref={topRef}
      >
        <div className="flex w-full flex-col items-center sm:mx-auto sm:max-w-xl">
          <h2 className="text-center text-2xl font-semibold">Completions</h2>
          <ul className="my-4 grid grid-cols-5 gap-4">
            <PracticeListItem num={1} completed={numCompleted >= 1} />
            <PracticeListItem num={2} completed={numCompleted >= 2} />
            <PracticeListItem num={3} completed={numCompleted >= 3} />
            <PracticeListItem num={4} completed={numCompleted >= 4} />
            <PracticeListItem num={5} completed={numCompleted >= 5} />
          </ul>
        </div>
        <div className="flex w-full flex-col items-center pt-4 sm:mx-auto sm:max-w-xl">
          <h2 className="text-center text-3xl font-semibold">How did it go?</h2>
          <div className="flex w-full flex-col justify-center gap-4 px-4 pt-8 xs:flex-row-reverse xs:px-0">
            <BigHappyButton
              disabled={!waitedLongEnough}
              onClick={succeed}
              className="sm:gap-2"
            >
              <span className="icon-[heroicons--hand-thumb-up-solid] -ml-1 size-6" />
              <span>Correct</span>
            </BigHappyButton>
            <BigAngryButton
              disabled={!waitedLongEnough}
              onClick={fail}
              className="sm:gap-2"
            >
              <span className="icon-[heroicons--hand-thumb-down-solid] -ml-1 size-6" />
              <span>Mistake</span>
            </BigAngryButton>
          </div>
        </div>
      </div>
      {spot && (
        <div className="px-8 pt-8">
          <PracticeSpotDisplay
            spot={spot}
            pieceid={pieceid}
            piecetitle={piecetitle}
          />
        </div>
      )}
      <div className="flex w-full flex-col py-12 sm:mx-auto sm:max-w-3xl">
        <div className="mx-auto flex w-full max-w-lg flex-wrap items-center justify-center">
          <WarningButton onClick={onFail}>
            <span>Move On</span>
            <span className="icon-[heroicons--arrow-up-right-solid] -ml-1 size-6" />
          </WarningButton>
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

function PracticeListItem({
  num,
  completed,
}: {
  num: number;
  completed: boolean;
}) {
  return (
    <AnimatePresence initial={false} mode="wait">
      {completed ? (
        <motion.li
          // @ts-ignore
          className="flex size-10 items-center justify-center rounded-xl border-2 border-green-700 bg-green-500/50 text-green-700 transition-all duration-100 sm:size-12"
          key={`${num}-completed`}
          initial="initial"
          animate="animate"
          exit="exit"
          variants={variants}
        >
          <span className="icon-[heroicons--check-solid] size-6 sm:size-8" />
          <span className="sr-only">Checked</span>
        </motion.li>
      ) : (
        <motion.li
          // @ts-ignore
          className="flex size-10 items-center justify-center rounded-xl border-2 border-neutral-700/10 bg-neutral-700/10 text-neutral-700/20 transition-all duration-100 sm:size-12"
          key={`${num}-incomplete`}
          initial="initial"
          animate="animate"
          exit="exit"
          variants={variants}
        >
          <div className="m-0 p-0 text-2xl font-bold ">{num}</div>
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
}: {
  restart: () => void;
  pieceid?: string;
  canPromote: boolean;
  promoteSpot: (toStage: string) => void;
  planid?: string;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  const close = useCallback(
    function () {
      if (dialogRef.current) {
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
    [dialogRef],
  );

  const handleRandom = useCallback(
    function () {
      promoteSpot("random");
      close();
    },
    [close, promoteSpot],
  );
  const handleMoreRepeat = useCallback(
    function () {
      promoteSpot("extra_repeat");
      close();
    },
    [close, promoteSpot],
  );

  useEffect(() => {
    if (dialogRef.current) {
      if (canPromote) {
        dialogRef.current.showModal();
      }
    }
  }, [canPromote, dialogRef.current]);

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
            Promote Spot
          </h3>
        </header>
        <div className="prose prose-sm prose-neutral mt-2 text-left">
          Now that you’ve repeat practiced your spot, it’s time to promote it to
          the next stage. Choose below whether you feel ready to move on to
          random practicing with this spot, or you feel it needs more repeat
          practicing.
        </div>
        <div className="mt-2 flex w-full flex-col gap-2 sm:flex-row-reverse sm:gap-2">
          <HappyButton
            grow
            onClick={handleRandom}
            className="h-14 w-full text-lg"
          >
            <span className="icon-[custom--shuffle] -ml-1 size-6"></span>
            Random
          </HappyButton>
          <SkyButton
            grow
            onClick={handleMoreRepeat}
            className="h-14 w-full text-lg"
          >
            <span className="icon-[custom--repeat] -ml-1 size-6"></span>
            More Repeat
          </SkyButton>
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
}: {
  restart: () => void;
  pieceid?: string;
  planid?: string;
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
}: {
  pieceid?: string;
  planid?: string;
  restart?: () => void;
}) {
  if (planid) {
    return (
      <>
        {!!pieceid && <BackToPiece pieceid={pieceid} />}
        <BackToPlan planid={planid} />
        <NextPlanItem planid={planid} />
      </>
    );
  }
  if (pieceid) {
    return (
      <>
        <BackToPiece pieceid={pieceid} />
        <WarningLink href={`/library/pieces/${pieceid}/practice/random-single`}>
          <span
            className="icon-[custom--shuffle] -ml-1 size-5"
            aria-hidden="true"
          ></span>
          Try Random Practicing
        </WarningLink>
        <HappyLink href={`/library/pieces/${pieceid}/practice/repeat`}>
          <span
            className="icon-[custom--repeat] -ml-1 size-5"
            aria-hidden="true"
          ></span>
          Practice Another Spot
        </HappyLink>
      </>
    );
  }
  return (
    <>
      <Link
        className="focusable action-button bg-sky-700/10 text-sky-800 hover:bg-sky-700/20"
        href="/library"
      >
        <span
          className="icon-[solar--music-note-slider-bold-duotone] -ml-1 size-5"
          aria-hidden="true"
        />
        Library
      </Link>
      <WarningLink href="/practice/random-single">
        <span
          className="icon-[custom--shuffle] -ml-1 size-5"
          aria-hidden="true"
        ></span>
        Try Random Practicing
      </WarningLink>
      <HappyButton onClick={restart}>
        <span
          className="icon-[custom--repeat] -ml-1 size-5"
          aria-hidden="true"
        ></span>
        Practice Another Spot
      </HappyButton>
    </>
  );
}
