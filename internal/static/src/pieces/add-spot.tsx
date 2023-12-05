import { yupResolver } from "@hookform/resolvers/yup";
import { useForm } from "react-hook-form";
import { SpotFormData, spotFormData } from "../validators";
import SpotFormFields from "./spot-form";
import { useState } from "preact/hooks";

export function AddSpotForm({
  pieceid,
  csrf,
  initialspotcount,
}: {
  pieceid: string;
  csrf: string;
  initialspotcount: string;
}) {
  const [isUpdating, setIsUpdating] = useState(false);
  const [numSpots, setNumSpots] = useState(parseInt(initialspotcount) || 0);
  const { handleSubmit, formState, setValue, reset, register, watch } =
    useForm<SpotFormData>({
      mode: "onBlur",
      reValidateMode: "onBlur",
      resolver: yupResolver(spotFormData),
      defaultValues: {
        name: "",
        idx: numSpots + 1,
        measures: "",
        audioPromptUrl: "",
        textPrompt: "",
        notesPrompt: "",
        imagePromptUrl: "",
        stage: "repeat",
      },
    });

  async function onSubmit(data: SpotFormData, e: Event) {
    e.preventDefault();
    setIsUpdating(true);
    // @ts-ignore
    await htmx.ajax("POST", `/library/pieces/${pieceid}/spots`, {
      values: data,
      target: "#spot-list",
      swap: "beforeend",
      headers: {
        "X-CSRF-Token": csrf,
      },
    });
    reset();
    setIsUpdating(false);
    document.getElementById("spot-count").textContent = `(${numSpots + 1})`;
    setNumSpots(numSpots + 1);
    setValue("idx", numSpots + 1);
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate hx-boost="false">
      <SpotFormFields
        csrf={csrf}
        isUpdating={isUpdating}
        formState={formState}
        setValue={setValue}
        backTo={`/library/pieces/${pieceid}`}
        register={register}
        watch={watch}
      />
    </form>
  );
}
