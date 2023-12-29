import { BasicSpot } from "../validators";
import {
  AudioPromptSummary,
  RemindersSummary,
  NotesPromptSummary,
  ImagePromptSummary,
} from "../ui/prompts";
import { cn } from "../common";

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
      <div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-neutral-500 bg-white/90 px-4 py-8 text-center text-3xl font-bold shadow-lg sm:px-8 sm:text-5xl md:col-span-2">
        {piecetitle && (
          <h4 className="-mb-3 -mt-2 text-lg text-neutral-700 underline underline-offset-1">
            {piecetitle}
          </h4>
        )}
        {spot.name ?? "Something went wrong"}
        {spot.measures && (
          <span className="-mb-2 text-lg text-neutral-700">
            Measures: {spot.measures}
          </span>
        )}
      </div>
      {shouldDisplayPrompts && (
        <div className="flex flex-col gap-2 rounded-xl border border-neutral-500 bg-white/90 px-4 pb-5 pt-4 shadow-lg sm:px-8 md:col-span-4">
          <h2 className="text-center text-lg font-semibold underline">
            Prompts
          </h2>
          <RemindersSummary
            text={spot.textPrompt}
            spotid={spot.id}
            pieceid={pieceid}
          />
          <AudioPromptSummary url={spot.audioPromptUrl} />
          <NotesPromptSummary notes={spot.notesPrompt} />
          <ImagePromptSummary url={spot.imagePromptUrl} />
        </div>
      )}
    </div>
  );
}
