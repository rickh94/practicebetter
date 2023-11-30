import { yupResolver } from "@hookform/resolvers/yup";
import { useForm } from "react-hook-form";
import { SpotFormData, spotFormData } from "../validators";
import SpotFormFields from "./spot-form";
import { useState } from "preact/hooks";

export function AddSpotForm({
  pieceid,
  csrf,
}: {
  pieceid: string;
  csrf: string;
}) {
  const [isUpdating, setIsUpdating] = useState(false);
  const { handleSubmit, formState, setValue, reset, register, watch } =
    useForm<SpotFormData>({
      mode: "onBlur",
      reValidateMode: "onBlur",
      resolver: yupResolver(spotFormData),
      defaultValues: {
        name: "",
        idx: 1,
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
