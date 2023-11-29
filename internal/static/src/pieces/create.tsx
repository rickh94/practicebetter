import { useForm } from "react-hook-form";
import { type PieceFormData, pieceFormData } from "../validators";
import { yupResolver } from "@hookform/resolvers/yup";
import { PieceFormFields } from "./piece-form";
// import htmx from "htmx.org";

// TODO: accept initial data as prop
export function CreatePieceForm({ csrf }: { csrf: string }) {
  const { register, control, handleSubmit, formState, watch, setValue } =
    useForm<PieceFormData>({
      mode: "onBlur",
      reValidateMode: "onChange",
      resolver: yupResolver(pieceFormData),
      defaultValues: {
        title: "",
        description: "",
        composer: "",
        practiceNotes: "",
        measures: null,
        goalTempo: null,
        beatsPerMeasure: null,
        spots: [],
      },
    });

  async function onSubmit(data: PieceFormData, e: Event) {
    e.preventDefault();
    // @ts-ignore
    await htmx.ajax("POST", "/library/pieces/create", {
      values: data,
      target: "#main-content",
      swap: "outerHTML",
      headers: {
        "X-CSRF-Token": csrf,
      },
    });
  }

  return (
    <form noValidate onSubmit={handleSubmit(onSubmit)} hx-boost="false">
      <PieceFormFields
        register={register}
        csrf={csrf}
        control={control}
        formState={formState}
        watch={watch}
        setValue={setValue}
      />
    </form>
  );
}
