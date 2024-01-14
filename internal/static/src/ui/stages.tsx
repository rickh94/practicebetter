import { getPieceStageDisplayName, getStageDisplayName } from "../common";

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
    <span className="flex items-center gap-1">
      {icon && <SpotStageIcon stage={stage} />}
      {getStageDisplayName(stage)}
    </span>
  );
}

export function SpotStageIcon({ stage }: { stage: string }) {
  switch (stage) {
    case "repeat":
      return (
        <span className="icon-[iconamoon--playlist-repeat-list-thin] mx-1 size-4" />
      );
    case "extra_repeat":
      return (
        <span className="icon-[iconamoon--playlist-repeat-list-thin] mx-1 size-4" />
      );
    case "random":
      return (
        <span className="icon-[iconamoon--playlist-shuffle-thin] mx-1 size-4" />
      );
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
