import {
  Control,
  FormState,
  UseFormRegister,
  UseFormWatch,
} from "react-hook-form";
import { PieceFormData } from "../vaidators";
import { FolderPlusIcon, XMarkIcon } from "@heroicons/react/20/solid";
import { HappyButton } from "../ui/buttons";
import { WarningLink } from "../ui/links";
import { SpotsArray } from "./spots-array";

export function PieceFormFields({
  control,
  register,
  formState,
  isUpdating = false,
  watch,
  backTo = "/library/pieces",
}: {
  control: Control<PieceFormData>;
  register: UseFormRegister<PieceFormData>;
  formState: FormState<PieceFormData>;
  isUpdating?: boolean;
  backTo?: string;
  watch: UseFormWatch<PieceFormData>;
}) {
  return (
    <>
      <div className="flex w-full flex-col">
        <div className="grid-cols-1 gap-x-0 py-2 sm:grid sm:gap-y-4 sm:px-0 md:grid-cols-5 md:gap-x-4">
          <div className="flex flex-col gap-1 sm:col-span-2">
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
          <div className="flex flex-col gap-1 sm:col-span-2">
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
          <div className="flex flex-col gap-1">
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
              {...register("goalTempo")}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.goalTempo && (
              <p className="text-sm text-red-600">
                {formState.errors.goalTempo.message}
              </p>
            )}
          </div>
        </div>
        <div className="grid-cols-1 py-2 sm:grid sm:grid-cols-2 sm:gap-4 sm:px-0 md:grid-cols-6">
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
              placeholder="100"
              {...register("measures")}
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
              Beats
            </label>
            <input
              type="number"
              id="beats"
              placeholder="Beats"
              {...register("beatsPerMeasure")}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.beatsPerMeasure && (
              <p className="text-sm text-red-600">
                {formState.errors.beatsPerMeasure.message}
              </p>
            )}
          </div>
          <div className="col-span-full flex w-full flex-col gap-1 sm:row-start-2 md:col-span-4 md:row-start-auto">
            <label
              className="text-sm font-medium leading-6 text-neutral-900"
              htmlFor="recording-link"
            >
              Recording Link
            </label>
            <input
              type="text"
              id="recording-link"
              placeholder="Reference Recording"
              {...register("recordingLink")}
              className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.recordingLink && (
              <p className="mt-2 text-sm text-red-600">
                {formState.errors.recordingLink.message}
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
              Practice Notes
            </label>
            <textarea
              id="practice-notes"
              placeholder="Things to remember"
              {...register("practiceNotes")}
              className="focusable h-24 w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
            />
            {formState.errors.practiceNotes && (
              <p className="mt-2 text-sm text-red-600">
                {formState.errors.practiceNotes.message}
              </p>
            )}
          </div>
        </div>
      </div>
      <SpotsArray
        control={control}
        register={register}
        formState={formState}
        watch={watch}
      />
      <div className="flex flex-row-reverse justify-start gap-4 py-4">
        <HappyButton disabled={!formState.isValid} type="submit">
          <FolderPlusIcon className="-ml-1 inline h-6 w-6" />
          {isUpdating ? "Saving..." : "Save"}
        </HappyButton>
        <WarningLink href={backTo}>
          <XMarkIcon className="-ml-1 inline h-6 w-6" />
          Cancel
        </WarningLink>
      </div>
    </>
  );
}
