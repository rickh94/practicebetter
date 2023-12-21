import {
  Control,
  FormState,
  UseFormRegister,
  UseFormSetValue,
  UseFormWatch,
} from "react-hook-form";
import { PieceFormData, pieceStages } from "../validators";
import { FolderPlusIcon, XMarkIcon } from "@heroicons/react/20/solid";
import { HappyButton } from "../ui/buttons";
import { WarningLink } from "../ui/links";
import { SpotsArray } from "./spots-array";
import { cn, getPieceStageDisplayName } from "../common";

// TODO: fix grid layout

export function PieceFormFields({
  control,
  register,
  formState,
  isUpdating = false,
  watch,
  backTo = "/library/pieces",
  csrf,
  setValue,
  showStage = false,
}: {
  control: Control<PieceFormData>;
  register: UseFormRegister<PieceFormData>;
  formState: FormState<PieceFormData>;
  isUpdating?: boolean;
  backTo?: string;
  watch: UseFormWatch<PieceFormData>;
  csrf: string;
  setValue: UseFormSetValue<PieceFormData>;
  showStage?: boolean;
}) {
  return (
    <>
      <div className="flex w-full flex-col">
        <div className="grid-cols-1 gap-x-0 py-2 sm:grid sm:gap-y-4 sm:px-0 md:grid-cols-2 md:gap-x-4">
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="title"
            >
              Title (required)
            </label>
            <input
              type="text"
              id="title"
              placeholder="Piece Title"
              {...register("title")}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.title && (
              <p className="text-sm text-red-600">
                {formState.errors.title.message}
              </p>
            )}
          </div>
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="composer"
            >
              Composer
            </label>
            <input
              type="text"
              id="composer"
              placeholder="Composer"
              {...register("composer")}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.composer && (
              <p className="text-sm text-red-600">
                {formState.errors.composer.message}
              </p>
            )}
          </div>
        </div>
        <div
          className={cn(
            "grid grid-cols-1 py-2",
            showStage
              ? "sm:grid-cols-2 sm:gap-4 sm:px-0 md:grid-cols-4"
              : "sm:grid-cols-2 sm:gap-4 sm:px-0 md:grid-cols-3",
          )}
        >
          <div
            className={cn(
              "flex flex-col gap-1 md:col-span-1",
              showStage ? "sm:col-span-1" : "sm:col-span-2",
            )}
          >
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="goal-tempo"
            >
              Goal Tempo
            </label>
            <input
              type="number"
              id="goal-tempo"
              placeholder="BPM"
              {...register("goalTempo", { valueAsNumber: true })}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.goalTempo && (
              <p className="text-sm text-red-600">
                {formState.errors.goalTempo.message}
              </p>
            )}
          </div>
          {showStage && (
            <div className={"flex flex-col gap-1 sm:col-span-1"}>
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
                {pieceStages.map((stage) => (
                  <option key={stage} value={stage}>
                    {getPieceStageDisplayName(stage)}
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
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="measures"
            >
              Measures
            </label>
            <input
              type="number"
              id="measures"
              placeholder="mm"
              {...register("measures", { valueAsNumber: true })}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.measures && (
              <p className="text-sm text-red-600">
                {formState.errors.measures.message}
              </p>
            )}
          </div>
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="beats"
            >
              Beats Per Measure
            </label>
            <input
              type="number"
              id="beats"
              placeholder="Beats"
              {...register("beatsPerMeasure", { valueAsNumber: true })}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.beatsPerMeasure && (
              <p className="text-sm text-red-600">
                {formState.errors.beatsPerMeasure.message}
              </p>
            )}
          </div>
        </div>
        <div className="grid-cols-1 py-2 sm:grid sm:gap-4 sm:px-0 md:grid-cols-2">
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="description"
            >
              Description
            </label>
            <textarea
              id="description"
              placeholder="Describe of your piece"
              {...register("description")}
              className="focusable h-24 w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.description && (
              <p className="mt-2 text-sm text-red-600">
                {formState.errors.description.message}
              </p>
            )}
          </div>
          <div className="flex flex-col gap-1">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="practice-notes"
            >
              {" "}
              Practice Notes{" "}
            </label>
            <textarea
              id="practice-notes"
              placeholder="Things to remember"
              {...register("practiceNotes")}
              className="focusable h-24 w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.practiceNotes && (
              <p className="mt-2 text-sm text-red-600">
                {" "}
                {formState.errors.practiceNotes.message}{" "}
              </p>
            )}
          </div>
        </div>
      </div>
      <SpotsArray
        setValue={setValue}
        csrf={csrf}
        control={control}
        register={register}
        formState={formState}
        watch={watch}
      />
      <div className="flex flex-row-reverse justify-start gap-4 py-4">
        <HappyButton type="submit">
          <FolderPlusIcon className="-ml-1 inline size-5" />
          {isUpdating ? "Saving..." : "Save"}
        </HappyButton>
        <WarningLink href={backTo}>
          <XMarkIcon className="-ml-1 inline size-5" />
          Cancel
        </WarningLink>
      </div>
    </>
  );
}
