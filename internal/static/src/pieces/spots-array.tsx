import { useAutoAnimate } from "@formkit/auto-animate/preact";
import { TrashIcon, PlusIcon } from "@heroicons/react/20/solid";
import {
  UseFormWatch,
  FormState,
  UseFormRegister,
  Control,
  useFieldArray,
  UseFieldArrayUpdate,
  UseFormSetValue,
} from "react-hook-form";
import { AngryButton } from "../ui/buttons";
import { PieceFormData, UpdatePieceData } from "../validators";
import {
  AddAudioPrompt,
  AddImagePrompt,
  AddTextPrompt,
  AddNotesPrompt,
} from "./add-prompts";

export function SpotsArray({
  control,
  register,
  formState,
  watch,
  csrf,
  setValue,
}: {
  watch: UseFormWatch<PieceFormData>;
  formState: FormState<PieceFormData>;
  register: UseFormRegister<PieceFormData>;
  control: Control<PieceFormData>;
  csrf: string;
  setValue: UseFormSetValue<PieceFormData>;
}) {
  const { fields, append, remove } = useFieldArray<
    PieceFormData | UpdatePieceData
  >({
    control,
    name: "spots",
  });

  const [parent] = useAutoAnimate();

  return (
    <>
      <h3 className="pt-2 text-left text-3xl font-bold">Add Practice Spots</h3>
      <ul
        ref={parent}
        className="grid grid-cols-1 gap-4 py-4 sm:grid-cols-2 lg:grid-cols-3"
      >
        {fields.map((item, index) => (
          <li
            key={item.id}
            className="flex flex-col justify-center gap-2 rounded-xl border border-neutral-500 bg-white/80 px-4 pb-4 pt-2 text-neutral-700"
          >
            <div className="flex flex-col">
              <label
                className="text-sm font-medium leading-6 text-neutral-900"
                htmlFor={`spots.${index}.name`}
              >
                Spot Name
              </label>
              <input
                id={`spots.${index}.name`}
                type="text"
                {...register(`spots.${index}.name`)}
                className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
              />
              {formState.errors.spots?.[index]?.name && (
                <p className="mt-2 text-sm text-red-600">
                  {formState.errors.spots?.[index]?.name?.message}
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              <div className="flex w-1/2 flex-col">
                <label
                  className="text-sm font-medium leading-6 text-neutral-900"
                  htmlFor={`spots.${index}.idx`}
                >
                  Spot Order
                </label>
                <input
                  id={`spots.${index}.order`}
                  type="number"
                  {...register(`spots.${index}.idx`)}
                  className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
                />
                {formState.errors.spots?.[index]?.idx && (
                  <p className="mt-2 text-sm text-red-600">
                    {formState.errors.spots?.[index]?.idx?.message}
                  </p>
                )}
              </div>
              <div className="flex w-1/2 flex-col">
                <label
                  className="text-sm font-medium leading-6 text-neutral-900"
                  htmlFor={`spots.${index}.measures`}
                >
                  Spot Measures
                </label>
                <input
                  id={`spots.${index}.measures`}
                  {...register(`spots.${index}.measures`)}
                  className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
                />
                {formState.errors.spots?.[index]?.measures && (
                  <p className="mt-2 text-sm text-red-600">
                    {formState.errors.spots?.[index]?.measures?.message}
                  </p>
                )}
              </div>
              <div className="flex w-1/2 flex-col">
                <label
                  className="text-sm font-medium leading-6 text-neutral-900"
                  htmlFor={`spots.${index}.currentTempo`}
                >
                  Current Tempo
                </label>
                <input
                  type="number"
                  placeholder="BPM"
                  id={`spots.${index}.currentTempo`}
                  {...register(`spots.${index}.currentTempo`)}
                  className="focusable w-full rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 placeholder-neutral-700 transition duration-200 focus:bg-neutral-700/20"
                />
                {formState.errors.spots?.[index]?.currentTempo && (
                  <p className="text-xs text-red-400">
                    {formState.errors.spots?.[index]?.currentTempo?.message}
                  </p>
                )}
              </div>
            </div>
            <div className="flex flex-col gap-1">
              <h4 className="text-sm font-medium leading-6 text-neutral-900">
                Prompts (optional)
              </h4>
              <div className="grid grid-cols-2 grid-rows-2 gap-2">
                <AddAudioPrompt
                  csrf={csrf}
                  save={(audioPromptUrl: string) =>
                    setValue(`spots.${index}.audioPromptUrl`, audioPromptUrl)
                  }
                  audioPromptUrl={watch(`spots.${index}.audioPromptUrl`)}
                />
                <AddImagePrompt
                  csrf={csrf}
                  save={(imagePromptUrl: string) =>
                    setValue(`spots.${index}.imagePromptUrl`, imagePromptUrl)
                  }
                  imagePromptUrl={watch(`spots.${index}.imagePromptUrl`)}
                />
                <AddTextPrompt
                  registerReturn={register(`spots.${index}.textPrompt`)}
                  textPrompt={watch(`spots.${index}.textPrompt`)}
                />
                <AddNotesPrompt
                  registerReturn={register(`spots.${index}.notesPrompt`)}
                  notesPrompt={watch(`spots.${index}.notesPrompt`)}
                />
              </div>
            </div>
            <AngryButton onClick={() => remove(index)}>
              <TrashIcon className="h-4 w-4" />
              Delete
            </AngryButton>
          </li>
        ))}
        <li className="flex min-h-[21rem] flex-col gap-2 rounded-xl border border-dashed border-neutral-500 bg-white/50 px-4 py-2 text-neutral-700 hover:bg-white/90 hover:text-black">
          <button
            className="flex h-full w-full items-center justify-center gap-1 text-2xl font-bold"
            type="button"
            onClick={() =>
              append({
                name: `Spot ${fields.length + 1}`,
                idx: fields.length + 1,
                stage: "repeat",
                measures: "mm 1-2",
                audioPromptUrl: "",
                textPrompt: "",
                notesPrompt: "",
                imagePromptUrl: "",
                currentTempo: null,
                id: null,
              })
            }
          >
            <PlusIcon className="h-6 w-6" />
            Add a Spot
          </button>
        </li>
      </ul>
    </>
  );
}
