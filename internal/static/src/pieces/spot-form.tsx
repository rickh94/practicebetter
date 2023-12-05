import { FolderPlusIcon } from "@heroicons/react/20/solid";
import {
  FormState,
  UseFormSetValue,
  UseFormWatch,
  UseFormRegister,
} from "react-hook-form";
import { cn, getStageDisplayName } from "../common";
import { HappyButton } from "../ui/buttons";
import { WarningLink } from "../ui/links";
import { SpotFormData, spotStages } from "../validators";
import {
  AddAudioPrompt,
  AddImagePrompt,
  AddTextPrompt,
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
}: {
  formState: FormState<SpotFormData>;
  setValue: UseFormSetValue<SpotFormData>;
  watch: UseFormWatch<SpotFormData>;
  register: UseFormRegister<SpotFormData>;
  isUpdating: boolean;
  backTo: string;
  showStage?: boolean;
  csrf: string;
}) {
  return (
    <>
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
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
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
              className="focusable block h-full w-full rounded-xl border-0 bg-neutral-700/10 py-2 pl-4 pr-12 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
              style={{
                appearance: "none",
                backgroundImage: `url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%23262626"><path fill-rule="evenodd" d="M12.53 16.28a.75.75 0 01-1.06 0l-7.5-7.5a.75.75 0 011.06-1.06L12 14.69l6.97-6.97a.75.75 0 111.06 1.06l-7.5 7.5z" clip-rule="evenodd" /></svg>')`,
                backgroundRepeat: "no-repeat",
                backgroundPosition: "right 0.7rem top 50%",
                backgroundSize: "1rem auto",
                WebkitAppearance: "none",
                textIndent: 1,
                textOverflow: "",
              }}
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
      <div className="flex gap-2">
        <div className="flex w-1/2 flex-col">
          <label
            className="text-sm font-medium leading-6 text-neutral-900"
            htmlFor="order"
          >
            Spot Order
          </label>
          <input
            type="number"
            id="idx"
            {...register("idx", { valueAsNumber: true })}
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
          />
          {formState.errors.idx && (
            <p className="text-xs text-red-400">
              {formState.errors.idx.message}
            </p>
          )}
        </div>
        <div className="flex w-1/2 flex-col">
          <label
            className="text-sm font-medium leading-6 text-neutral-900"
            htmlFor="measures"
          >
            Measures
          </label>
          <input
            id="measures"
            {...register("measures")}
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
          />
          {formState.errors.measures && (
            <p className="text-xs text-red-400">
              {formState.errors.measures.message}
            </p>
          )}
        </div>
        <div className="flex w-1/2 flex-col">
          <label
            className="text-sm font-medium leading-6 text-neutral-900"
            htmlFor="current-tempo"
          >
            Curr Tempo
          </label>
          <input
            {...register("currentTempo", { valueAsNumber: true })}
            type="number"
            placeholder="BPM"
            id="current-tempo"
            className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
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
        <div className="grid grid-cols-2 grid-rows-2 gap-2 md:grid-cols-4 md:grid-rows-1 lg:grid-cols-2 lg:grid-rows-1">
          <AddAudioPrompt
            csrf={csrf}
            save={(url) => setValue("audioPromptUrl", url)}
            audioPromptUrl={watch("audioPromptUrl")}
          />
          <AddImagePrompt
            csrf={csrf}
            save={(url) => setValue("imagePromptUrl", url)}
            imagePromptUrl={watch("imagePromptUrl")}
          />
          <AddTextPrompt
            textPrompt={watch("textPrompt")}
            registerReturn={register("textPrompt")}
          />
          <AddNotesPrompt
            registerReturn={register("notesPrompt")}
            notesPrompt={watch("notesPrompt")}
          />
        </div>
      </div>
      <div className="flex flex-row-reverse justify-start gap-4 pt-4">
        <HappyButton disabled={!formState.isValid} type="submit">
          <FolderPlusIcon className="-ml-1 h-6 w-6" />
          {isUpdating ? "Saving..." : "Save"}
        </HappyButton>
        <WarningLink href={backTo}>← Go Back</WarningLink>
      </div>
    </>
  );
}
