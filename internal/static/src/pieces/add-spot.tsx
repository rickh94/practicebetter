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
  const [nextSpotIdx, setNextSpotIdx] = useState(
    parseInt(initialspotcount) + 1 || 1,
  );
  const { handleSubmit, formState, setValue, reset, register, watch } =
    useForm<SpotFormData>({
      mode: "onBlur",
      reValidateMode: "onBlur",
      resolver: yupResolver(spotFormData),
      defaultValues: {
        name: "",
        idx: nextSpotIdx,
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
    document.getElementById("spot-count").textContent = `(${nextSpotIdx})`;
    setNextSpotIdx(nextSpotIdx + 1);
    setValue("idx", nextSpotIdx + 1);
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      noValidate
      hx-boost="false"
      id="add-spot-form"
    >
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
