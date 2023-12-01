import { yupResolver } from "@hookform/resolvers/yup";
import { useForm } from "react-hook-form";
import { SpotFormData, spotFormData } from "../validators";
import SpotFormFields from "./spot-form";
import { useState } from "preact/hooks";

export function EditSpotForm({
  pieceid,
  csrf,
  spotdata,
  spotid,
}: {
  pieceid: string;
  csrf: string;
  spotdata: string;
  spotid: string;
}) {
  const [isUpdating, setIsUpdating] = useState(false);
  const { handleSubmit, formState, setValue, reset, register, watch } =
    useForm<SpotFormData>({
      mode: "onBlur",
      reValidateMode: "onBlur",
      resolver: yupResolver(spotFormData),
      defaultValues: async () => {
        const spot: SpotFormData = JSON.parse(spotdata);
        return {
          name: spot.name ?? "",
          idx: spot.idx ?? 1,
          measures: spot.measures ?? "",
          audioPromptUrl: spot.audioPromptUrl ?? "",
          textPrompt: spot.textPrompt ?? "",
          notesPrompt: spot.notesPrompt ?? "",
          imagePromptUrl: spot.imagePromptUrl ?? "",
          stage: spot.stage ?? "repeat",
        };
      },
    });

  async function onSubmit(data: SpotFormData, e: Event) {
    e.preventDefault();
    setIsUpdating(true);
    // @ts-ignore
    await htmx.ajax("PUT", `/library/pieces/${pieceid}/spots/${spotid}`, {
      values: data,
      target: "#main-content",
      swap: "outerHTML",
      headers: {
        "X-CSRF-Token": csrf,
      },
    });
    reset();
    setIsUpdating(false);
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate hx-boost="false">
      <SpotFormFields
        csrf={csrf}
        isUpdating={isUpdating}
        formState={formState}
        setValue={setValue}
        backTo={`/library/pieces/${pieceid}/spots/${spotid}`}
        register={register}
        watch={watch}
        showStage
      />
    </form>
  );
}
