import {
  type FormState,
  type UseFormSetValue,
  type UseFormWatch,
  type UseFormRegister,
} from "react-hook-form";
import { cn, getStageDisplayName } from "../common";
import { WarningLink } from "../ui/links";
import { type SpotFormData, spotStages } from "../validators";
import {
  AddAudioPrompt,
  AddImagePrompt,
  AddReminders,
  AddNotesPrompt,
} from "./add-prompts";

export default function SpotFormFields({
  formState,
  isUpdating,
  setValue,
  watch,
  register,
  backTo,
  showStage = false,
  csrf,
  spotid = "",
  pieceid = "",
}: {
  formState: FormState<SpotFormData>;
  setValue: UseFormSetValue<SpotFormData>;
  watch: UseFormWatch<SpotFormData>;
  register: UseFormRegister<SpotFormData>;
  isUpdating: boolean;
  backTo: string;
  showStage?: boolean;
  csrf: string;
  spotid?: string;
  pieceid?: string;
}) {
  return (
    <div>
      <div className="grid grid-cols-2 gap-2">
        <div
          className={cn(
            "flex flex-col",
            showStage ? "col-span-1" : "col-span-full",
          )}
        >
          <label
            className="text-sm font-medium leading-6 text-neutral-900"
            htmlFor="name"
          >
            Spot Name
          </label>
          <input
            {...register("name")}
            id="name"
            placeholder="Spot #1"
            className="basic-field w-full"
            autoComplete="off"
          />
          {formState.errors.name && (
            <p className="text-xs text-red-400">
              {formState.errors.name.message}
            </p>
          )}
        </div>
        {showStage && (
          <div className="col-span-1 flex flex-col">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="stage"
            >
              Stage
            </label>
            <select
              {...register("stage")}
              id="stage"
              className="basic-field custom-select h-full w-full"
            >
              {spotStages.map((stage) => (
                <option key={stage} value={stage}>
                  {getStageDisplayName(stage)}
                </option>
              ))}
            </select>
            {formState.errors.stage && (
              <p className="text-xs text-red-400">
                {formState.errors.stage.message}
              </p>
            )}
          </div>
        )}
      </div>
      <div className="flex w-full gap-2">
        <div className="flex flex-grow flex-col">
          <label
            className="text-sm font-medium leading-6 text-neutral-900"
            htmlFor="measures"
          >
            Measures
          </label>
          <input
            id="measures"
            placeholder="1-3"
            {...register("measures")}
            className="basic-field w-full"
          />
          {formState.errors.measures && (
            <p className="text-xs text-red-400">
              {formState.errors.measures.message}
            </p>
          )}
        </div>
        <div className="flex flex-grow flex-col">
          <label
            className="text-sm font-medium leading-6 text-neutral-900"
            htmlFor="current-tempo"
          >
            <span className="hidden xs:inline">Current </span>
            Tempo
          </label>
          <input
            {...register("currentTempo", { valueAsNumber: true })}
            type="number"
            placeholder="BPM"
            id="current-tempo"
            className="basic-field w-full"
          />
          {formState.errors.currentTempo && (
            <p className="text-xs text-red-400">
              {formState.errors.currentTempo.message}
            </p>
          )}
        </div>
      </div>
      <div className="mt-2 flex flex-col gap-2">
        <div>
          <h4 className="text-lg font-medium leading-6 text-neutral-900">
            Prompts (optional)
          </h4>
          <p className="text-sm italic">
            Add small prompts to help you play this spot correctly
          </p>
        </div>
        <div className="grid grid-cols-2 grid-rows-2 gap-2">
          <AddAudioPrompt
            csrf={csrf}
            save={(url) => setValue("audioPromptUrl", url)}
            audioPromptUrl={watch("audioPromptUrl")}
            spotid={spotid}
            pieceid={pieceid}
          />
          <AddImagePrompt
            csrf={csrf}
            save={(url) => setValue("imagePromptUrl", url)}
            imagePromptUrl={watch("imagePromptUrl")}
            spotid={spotid}
            pieceid={pieceid}
          />
          <AddReminders
            textPrompt={watch("textPrompt")}
            registerReturn={register("textPrompt")}
          />
          <AddNotesPrompt
            registerReturn={register("notesPrompt")}
            notesPrompt={watch("notesPrompt")}
          />
        </div>
      </div>
      <div
        className="flex flex-row-reverse justify-start gap-2 pt-4"
        id="spot-form-button-row"
      >
        <button type="submit" className="action-button green focusable">
          <span
            className="icon-[iconamoon--arrow-up-5-circle-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          {isUpdating ? "Saving..." : "Save"}
        </button>
        <WarningLink href={backTo}>
          <span
            className="icon-[iconamoon--arrow-left-5-circle-thin] -ml-1 size-5"
            aria-hidden="true"
          />
          Go Back
        </WarningLink>
      </div>
    </div>
  );
}
