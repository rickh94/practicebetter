import { useForm } from "react-hook-form";
import { type PieceFormData, pieceFormData } from "../validators";
import { yupResolver } from "@hookform/resolvers/yup";
import { PieceFormFields } from "./piece-form";
// import htmx from "htmx.org";

export function EditPieceForm({
  csrf,
  piece,
  spots,
  pieceid,
}: {
  csrf: string;
  piece: string;
  spots: string;
  pieceid: string;
}) {
  const initialPieceInfo = JSON.parse(piece);
  const initialSpots: any[] = JSON.parse(spots);
  const { register, control, handleSubmit, formState, watch } =
    useForm<PieceFormData>({
      mode: "onBlur",
      reValidateMode: "onChange",
      resolver: yupResolver(pieceFormData),
      defaultValues: {
        title: initialPieceInfo.title ?? "",
        description: initialPieceInfo.description ?? "",
        composer: initialPieceInfo.composer ?? "",
        practiceNotes: initialPieceInfo.practiceNotes ?? "",
        measures: initialPieceInfo.measures ?? null,
        spots: initialSpots.map((spot, idx) => ({
          id: spot.id ?? "",
          name: spot.name ?? "",
          idx: spot.idx ?? idx,
          stage: spot.stage ?? "repeat",
          measures: spot.measures ?? "",
          audioPromptUrl: spot.audioPromptUrl ?? "",
          imagePromptUrl: spot.imagePromptUrl ?? "",
          notesPrompt: spot.notesPrompt ?? "",
          textPrompt: spot.textPrompt ?? "",
          currentTempo: spot.currentTempo ?? null,
        })),
      },
    });

  async function onSubmit(data: PieceFormData, e: Event) {
    e.preventDefault();
    // @ts-ignore
    await htmx.ajax("PUT", `/library/pieces/${pieceid}`, {
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
        control={control}
        formState={formState}
        watch={watch}
        backTo={`/library/pieces/${pieceid}`}
      />
    </form>
  );
}
