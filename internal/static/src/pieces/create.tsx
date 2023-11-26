import { useForm } from "react-hook-form";
import { type PieceFormData, pieceFormData } from "../vaidators";
import { zodResolver } from "@hookform/resolvers/zod";
import { PieceFormFields } from "./piece-form";
import htmx from "htmx.org";

// TODO: accept initial data as prop
export function CreatePieceForm({ csrf }: { csrf: string }) {
  const { register, control, handleSubmit, formState, watch } =
    useForm<PieceFormData>({
      mode: "onBlur",
      reValidateMode: "onChange",
      resolver: zodResolver(pieceFormData),
      defaultValues: {
        title: "",
        description: "",
        composer: "",
        recordingLink: "",
        practiceNotes: "",
        spots: [],
      },
    });

  async function onSubmit(data: PieceFormData) {
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
    <form noValidate onSubmit={handleSubmit(onSubmit)}>
      <PieceFormFields
        register={register}
        control={control}
        formState={formState}
        watch={watch}
      />
    </form>
  );
}
