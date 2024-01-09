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
      return (
        <span
          className="icon-[iconamoon--bookmark-thin] mx-1 size-4"
          aria-hidden="true"
        />
      );
    case "interleave_days":
      return (
        <span
          className="icon-[iconamoon--calendar-1-thin] mx-1 size-4"
          aria-hidden="true"
        />
      );
    case "completed":
      return (
        <span
          className="icon-[iconamoon--check-circle-1-duotone] mx-1 size-4"
          aria-hidden="true"
        />
      );
    default:
      return <></>;
  }
}
