import { getPieceStageDisplayName, getStageDisplayName } from "../common";

export function PieceStage({ stage }: { stage: string }) {
  return <>{getPieceStageDisplayName(stage)}</>;
}

export function SpotStage({ stage }: { stage: string }) {
  return <>{getStageDisplayName(stage)}</>;
}
