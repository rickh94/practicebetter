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
  const { register, control, handleSubmit, formState, watch, setValue } =
    useForm<PieceFormData>({
      mode: "onBlur",
      reValidateMode: "onChange",
      resolver: yupResolver(pieceFormData),
      defaultValues: async () => {
        const initialPieceInfo = JSON.parse(piece);
        const initialSpots: any[] = JSON.parse(spots) ?? [];
        return {
          id: initialPieceInfo.id ?? "",
          title: initialPieceInfo.title ?? "",
          description: initialPieceInfo.description ?? "",
          composer: initialPieceInfo.composer ?? "",
          practiceNotes: initialPieceInfo.practiceNotes ?? "",
          measures: initialPieceInfo.measures ?? undefined,
          goalTempo: initialPieceInfo.goalTempo ?? undefined,
          beatsPerMeasure: initialPieceInfo.beatsPerMeasure ?? undefined,
          stage: initialPieceInfo.stage ?? "active",
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
            currentTempo: spot.currentTempo ?? undefined,
          })),
        };
      },
    });

  async function onSubmit(data: PieceFormData, e: Event) {
    e.preventDefault();
    console.log(data);
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
        csrf={csrf}
        register={register}
        control={control}
        formState={formState}
        watch={watch}
        backTo={`/library/pieces/${pieceid}`}
        setValue={setValue}
        showStage={true}
      />
    </form>
  );
}
