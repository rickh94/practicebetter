import {
  BookmarkIcon,
  CalendarDaysIcon,
  CheckCircleIcon,
} from "@heroicons/react/20/solid";
import { getPieceStageDisplayName, getStageDisplayName } from "../common";
import { RepeatIcon, ShuffleIcon } from "./icons";

export function PieceStage({ stage }: { stage: string }) {
  return <>{getPieceStageDisplayName(stage)}</>;
}

export function SpotStage({
  stage,
  icon = false,
}: {
  stage: string;
  icon?: boolean;
}) {
  return (
    <>
      {icon && <SpotStageIcon stage={stage} />}
      {getStageDisplayName(stage)}
    </>
  );
}

export function SpotStageIcon({ stage }: { stage: string }) {
  switch (stage) {
    case "repeat":
      return <RepeatIcon className="mx-1 inline size-4" />;
    case "extra_repeat":
      return <RepeatIcon className="mx-1 inline size-4" />;
    case "random":
      return <ShuffleIcon className="mx-1 inline size-4" />;
    case "interleave":
      return <BookmarkIcon className="mx-1 inline size-4" />;
    case "interleave_days":
      return <CalendarDaysIcon className="mx-1 inline size-4" />;
    case "completed":
      return <CheckCircleIcon className="mx-1 inline size-4" />;
    default:
      return <></>;
  }
}
