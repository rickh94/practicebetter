import { type BasicSpot } from "../validators";
import {
  AudioPromptSummary,
  RemindersSummary,
  NotesPromptSummary,
  ImagePromptSummary,
} from "../ui/prompts";
import { cn } from "../common";

function getSpotNameTextSize(spot: BasicSpot) {
  const length = spot?.name?.length ?? 20;
  if (length < 10) {
    return "text-4xl";
  }
  if (length < 20) {
    return "text-3xl";
  }
  return "text-xl";
}

// TODO:: set some min heights so the layout doesn't shift around as much
export function PracticeSpotDisplay({
  spot,
  pieceid = "",
  piecetitle = "",
}: {
  spot: BasicSpot;
  pieceid?: string;
  piecetitle?: string;
}) {
  if (!spot) {
    return <>Missing Spot data</>;
  }
  const shouldDisplayPrompts =
    !!pieceid ||
    !!spot.audioPromptUrl ||
    !!spot.textPrompt ||
    !!spot.imagePromptUrl ||
    !!spot.notesPrompt;
  return (
    <div
      className={cn(
        shouldDisplayPrompts
          ? "mx-auto grid max-w-md gap-2 md:mx-0 md:max-w-full md:grid-cols-6"
          : "flex flex-col items-center justify-center",
      )}
    >
      <div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-neutral-500 bg-white px-4 py-8 text-center font-bold shadow-lg sm:px-8 md:col-span-2">
        {piecetitle && (
          <h4 className="-mb-3 -mt-2 text-sm italic text-neutral-900">
            {piecetitle}
          </h4>
        )}
        <span className={cn("text-pretty", getSpotNameTextSize(spot))}>
          {spot.name ?? "Something went wrong"}
        </span>
        {spot.measures && (
          <span className="-mb-2 text-lg text-neutral-700">
            Measures: {spot.measures}
          </span>
        )}
      </div>
      {shouldDisplayPrompts && (
        <div className="flex flex-col gap-2 rounded-xl border border-neutral-500 bg-white px-4 pb-5 pt-4 shadow-lg sm:px-8 md:col-span-4">
          <h2 className="text-center text-lg font-semibold underline">
            Prompts
          </h2>
          <RemindersSummary
            text={spot.textPrompt ?? ""}
            spotid={spot.id ?? undefined}
            pieceid={pieceid}
          />
          <AudioPromptSummary url={spot.audioPromptUrl ?? ""} />
          <NotesPromptSummary notes={spot.notesPrompt ?? ""} />
          <ImagePromptSummary url={spot.imagePromptUrl ?? ""} />
        </div>
      )}
    </div>
  );
}

export function PracticeSpotDisplayWrapper({
  spotjson = "",
  pieceid = "",
  piecetitle = "",
}: {
  spotjson: string;
  pieceid?: string;
  piecetitle?: string;
}) {
  const spot = JSON.parse(spotjson) as BasicSpot;
  if (!spot?.name) {
    return <>Missing Spot data</>;
  }
  return (
    <PracticeSpotDisplay
      spot={spot}
      pieceid={pieceid}
      piecetitle={piecetitle}
    />
  );
}
