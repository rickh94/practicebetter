import { yupResolver } from "@hookform/resolvers/yup";
import { useForm } from "react-hook-form";
import { type SpotFormData, spotFormData } from "../validators";
import SpotFormFields from "./spot-form";
import { useState } from "preact/hooks";
import * as htmx from "htmx.org";

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
      // eslint-disable-next-line @typescript-eslint/require-await
      defaultValues: async () => {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        const spot: SpotFormData = JSON.parse(spotdata);
        return {
          name: spot.name ?? "",
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
    await htmx.ajax("PUT", `/library/pieces/${pieceid}/spots/${spotid}`, {
      values: data,
      target: "#main-content",
      swap: "outerHTML transition:true",
      headers: { "X-CSRF-Token": csrf },
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
        spotid={spotid}
        pieceid={pieceid}
      />
    </form>
  );
}
