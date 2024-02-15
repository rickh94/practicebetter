import { useCallback, useEffect, useRef, useState } from "preact/hooks";
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

// TODO: start/cancel editing callback options to disable buttons and such
// TODO:: set some min heights so the layout doesn't shift around as much
export function PracticeSpotDisplay({
  onEdit,
  spot,
  pieceid = "",
  piecetitle = "",
  csrf,
  updateSpot,
}: {
  spot: BasicSpot;
  pieceid?: string;
  piecetitle?: string;
  onEdit?: (data: BasicSpot) => void;
  csrf?: string;
  updateSpot?: (data: BasicSpot) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const editFormRef = useRef(null);

  const startEditing = useCallback(() => {
    setIsEditing(true);
  }, []);

  const stopEditing = useCallback(() => {
    setIsEditing(false);
  }, []);

  const save = useCallback(
    async (e: Event) => {
      e.preventDefault();
      if (!editFormRef.current) {
        return;
      }
      if (!pieceid || !spot?.id) {
        return;
      }
      const body = new FormData(editFormRef.current);
      try {
        const res = await fetch(`/library/pieces/${pieceid}/spots/${spot.id}`, {
          method: "PATCH",
          body,
        });
        if (res.ok) {
          const spot = (await res.json()) as BasicSpot;
          updateSpot?.(spot);
          setIsEditing(false);
          onEdit?.(spot);
        } else {
          const error = await res.text();
          setError(error);
        }
      } catch (e) {
        console.log(e);
        setError("Something went wrong");
      }
    },
    [onEdit, pieceid, spot.id, updateSpot],
  );

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
    <>
      {isEditing ? (
        <form
          className={cn(
            shouldDisplayPrompts
              ? "mx-auto flex max-w-md flex-col gap-2"
              : "flex flex-col items-center justify-center",
          )}
          ref={editFormRef}
          onSubmit={save}
        >
          <input type="hidden" name="gorilla.csrf.Token" value={csrf} />
          <div className="relative flex flex-col items-center justify-center gap-2 rounded-xl border border-neutral-500 bg-white p-4 text-center font-bold shadow-lg">
            {piecetitle && (
              <h4 className="-mb-1 -mt-2 text-sm italic text-neutral-900">
                {piecetitle}
              </h4>
            )}
            <div class="flex w-full flex-col gap-1 pt-2">
              <label
                className="flex-grow-0 text-left text-sm font-semibold"
                htmlFor="name"
              >
                Name
              </label>
              <input
                value={spot.name ?? ""}
                className="basic-field neutral focusable min-w-0 flex-grow"
                name="name"
                type="text"
                placeholder="Spot Name"
                id="name"
              />
            </div>
            <div className="flex w-full flex-col gap-1">
              <label
                htmlFor="currentTempo"
                className="flex-grow-0 text-left text-sm font-semibold"
              >
                Current Tempo
              </label>
              <input
                className="basic-field neutral focusable min-w-0 flex-grow"
                name="currentTempo"
                value={spot.currentTempo ?? ""}
                type="text"
                placeholder="BPM"
                id="currentTempo"
              />
            </div>
            <div className="flex w-full flex-col gap-1">
              <label
                htmlFor="measures"
                className="flex-grow-0 text-left text-sm font-semibold"
              >
                Measures
              </label>
              <input
                className="basic-field neutral focusable min-w-0 flex-grow"
                name="measures"
                value={spot.measures ?? ""}
                type="text"
                placeholder="mm1-2"
                id="measures"
              />
            </div>
            {error ? (
              <div className="flex w-full flex-col justify-start gap-1">
                <p className="text-left italic text-red-500">{error}</p>
              </div>
            ) : null}
            <div className="mt-2 flex w-full flex-col flex-wrap items-center justify-start gap-2 xs:flex-row-reverse">
              <button
                type="submit"
                className="action-button green focusable w-full xs:w-auto"
              >
                <span
                  className="icon-[iconamoon--arrow-up-5-circle-thin] -ml-1 size-6"
                  aria-hidden="true"
                />
                Save
              </button>
              <button
                type="button"
                onClick={stopEditing}
                className="action-button red focusable w-full xs:w-auto"
              >
                <span
                  className="icon-[iconamoon--sign-times-circle-thin] -ml-1 size-6"
                  aria-hidden="true"
                />
                Cancel
              </button>
            </div>
          </div>
          {shouldDisplayPrompts && (
            <PracticeSpotPrompts
              spot={spot}
              pieceid={pieceid}
              csrf={csrf}
              updateSpot={updateSpot}
            />
          )}
        </form>
      ) : (
        <div
          className={cn(
            shouldDisplayPrompts
              ? "mx-auto grid max-w-md gap-2 md:mx-0 md:max-w-full md:grid-cols-6"
              : "flex flex-col items-center justify-center",
          )}
        >
          <div className="relative flex flex-col items-center justify-center gap-2 rounded-xl border border-neutral-500 bg-white px-4 py-8 text-center font-bold shadow-lg sm:px-8 md:col-span-2">
            {csrf ? (
              <button
                onClick={startEditing}
                className="focusable neutral group absolute right-0 top-0 pb-0 pl-2 pr-3 pt-2 font-medium italic"
              >
                <span className="flex items-center gap-1 border-neutral-500 px-1 group-hover:border-b">
                  Edit
                  <span
                    className="icon-[iconamoon--edit-thin] -mr-1"
                    aria-hidden="true"
                  />
                </span>
              </button>
            ) : null}
            {piecetitle && (
              <h4 className="-mb-1 -mt-2 text-sm italic text-neutral-900">
                {piecetitle}
              </h4>
            )}
            <span className={cn("text-pretty", getSpotNameTextSize(spot))}>
              {spot.name ?? "Something went wrong"}
            </span>
            {spot.currentTempo && (
              <span className="mt-2 text-base font-medium text-neutral-700">
                Current Tempo: {spot.currentTempo}
              </span>
            )}
            {spot.measures && (
              <span className="-mb-2 text-lg text-neutral-700">
                Measures: {spot.measures}
              </span>
            )}
          </div>
          {shouldDisplayPrompts && (
            <PracticeSpotPrompts
              spot={spot}
              pieceid={pieceid}
              csrf={csrf}
              updateSpot={updateSpot}
            />
          )}
        </div>
      )}
    </>
  );
}

function PracticeSpotPrompts(props: {
  spot: BasicSpot;
  pieceid?: string;
  csrf?: string;
  updateSpot?: (spot: BasicSpot) => void;
}) {
  const saveReminders = useCallback(
    (text: string) => {
      props.updateSpot?.({ ...props.spot, textPrompt: text });
    },
    [props],
  );

  const saveImage = useCallback(
    (url: string) => {
      props.updateSpot?.({ ...props.spot, imagePromptUrl: url });
    },
    [props],
  );

  const saveAudio = useCallback(
    (url: string) => {
      props.updateSpot?.({ ...props.spot, audioPromptUrl: url });
    },
    [props],
  );

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-neutral-500 bg-white px-4 pb-5 pt-4 shadow-lg md:col-span-4">
      <h2 className="text-center text-lg font-semibold underline">Prompts</h2>
      <RemindersSummary
        text={props.spot.textPrompt ?? ""}
        spotid={props.spot.id ?? undefined}
        pieceid={props.pieceid}
        csrf={props.csrf}
        save={saveReminders}
      />
      <ImagePromptSummary
        url={props.spot.imagePromptUrl ?? ""}
        spotid={props.spot.id ?? ""}
        pieceid={props.pieceid}
        csrf={props.csrf}
        save={saveImage}
      />
      <AudioPromptSummary
        url={props.spot.audioPromptUrl ?? ""}
        spotid={props.spot.id ?? ""}
        pieceid={props.pieceid}
        csrf={props.csrf}
        save={saveAudio}
      />
      <NotesPromptSummary notes={props.spot.notesPrompt ?? ""} />
    </div>
  );
}

export function PracticeSpotDisplayWrapper({
  spotjson = "",
  pieceid = "",
  piecetitle = "",
  csrf = "",
}: {
  spotjson: string;
  pieceid?: string;
  piecetitle?: string;
  csrf?: string;
}) {
  const [spot, setSpot] = useState<BasicSpot | null>(null);

  useEffect(() => {
    const initialspot = JSON.parse(spotjson) as BasicSpot;
    if (!initialspot?.name) {
      return;
    }
    setSpot(initialspot);
  }, [spotjson]);

  const updateSpot = useCallback((spot: BasicSpot) => {
    setSpot(spot);
  }, []);

  return (
    <>
      {spot ? (
        <PracticeSpotDisplay
          spot={spot}
          pieceid={pieceid}
          piecetitle={piecetitle}
          updateSpot={updateSpot}
          csrf={csrf}
        />
      ) : (
        <div>Missing spot info</div>
      )}
    </>
  );
}
