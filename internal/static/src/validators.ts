import * as yup from "yup";

const optionalPosInt = yup
  .number()
  .positive()
  .integer()
  .optional()
  .nullable()
  .transform((_, val: number | undefined | null) =>
    val === Number(val) ? val : null,
  );

export const spotStages = [
  "repeat",
  "extra_repeat",
  "random",
  "interleave",
  "interleave_days",
  "completed",
] as const;
export const spotStage = yup.string().oneOf(spotStages);

export const pieceStages = ["active", "future", "completed"] as const;
export const pieceStage = yup.string().oneOf(pieceStages);

export const basicSpot = yup.object({
  id: yup.string().nullable().optional(),
  name: yup.string().min(1, "Too Short"),
  stage: spotStage.default("repeat"),
  measures: yup.string().default("").optional(),
  audioPromptUrl: yup.string().nullable().optional(),
  imagePromptUrl: yup.string().nullable().optional(),
  notesPrompt: yup.string().nullable().optional(),
  textPrompt: yup.string().nullable().optional(),
  currentTempo: optionalPosInt,
  stageStarted: optionalPosInt,
});

export const spotFormData = basicSpot;

export const pieceWithSpots = yup.object({
  id: yup.string().nullable().optional(),
  title: yup
    .string()
    .min(1, "Title must be at least one letter")
    .max(255, "Title is too long."),
  description: yup.string().optional(),
  composer: yup.string().optional(),
  measures: optionalPosInt,
  beatsPerMeasure: optionalPosInt,
  practiceNotes: yup.string().optional(),
  goalTempo: optionalPosInt,
  spots: yup.array(basicSpot),
  stage: pieceStage.default("active"),
});

export const pieceFormData = pieceWithSpots;

export const basicPiece = pieceWithSpots.omit(["spots"]);

export const updatePieceWithSpots = pieceFormData;
export type BasicPiece = yup.InferType<typeof basicPiece>;
export type PieceWithSpots = yup.InferType<typeof pieceWithSpots>;
export type UpdatePieceData = yup.InferType<typeof updatePieceWithSpots>;
export type PieceFormData = yup.InferType<typeof pieceFormData>;

export type BasicSpot = yup.InferType<typeof basicSpot>;
export type SpotFormData = yup.InferType<typeof spotFormData>;
export type SpotStage = yup.InferType<typeof spotStage>;
