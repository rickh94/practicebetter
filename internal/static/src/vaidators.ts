import { z } from "zod";

export const spotStage = z.enum([
  "repeat",
  "random",
  "interleave",
  "interleave_days",
  "completed",
]);

export const urlOrEmpty = z.union([z.string().url(), z.enum([""])]);

export const urlOrEmptyForm = z.object({ url: urlOrEmpty });

export const notesForm = z.object({
  notes: z.string(),
});

export const textForm = z.object({
  text: z.string(),
});

export const basicSpot = z.object({
  id: z.string(),
  name: z.string().min(1, "Too Short"),
  order: z.coerce.number().nullish().optional(),
  stage: spotStage.default("repeat"),
  measures: z.string().default("").optional(),
  audioPromptUrl: z
    .union([z.string().url(), z.null(), z.enum([""])])
    .optional(),
  imagePromptUrl: z
    .union([z.string().url(), z.null(), z.enum([""])])
    .optional(),
  notesPrompt: z.string().nullable().optional(),
  textPrompt: z.string().nullable().optional(),
  currentTempo: z.union([z.coerce.number().min(1), z.null()]).optional(),
});

// React hook form gets mad if you pass it something nullable, so I omit the order field and re-add it as optional

export const spotFormData = basicSpot
  .omit({ id: true, currentTempo: true })
  .extend({
    id: z.string().optional(),
    currentTempo: z.coerce.number().nullish().optional(),
  });

export const spotWithPieceInfo = basicSpot.extend({
  piece: z.object({
    title: z.string(),
    id: z.string(),
  }),
});

export const pieceWithSpots = z.object({
  id: z.string(),
  title: z
    .string()
    .min(1, "Title must be at least one letter")
    .max(255, "Title is too long."),
  description: z.string(),
  composer: z.string(),
  recordingLink: z.union([z.string().url(), z.enum([""]), z.null()]),
  measures: z.coerce.number().nullish(),
  beatsPerMeasure: z.coerce.number().nullish(),
  practiceNotes: z.string().nullish(),
  goalTempo: z.coerce.number().nullish(),
  spots: z.array(basicSpot),
});

export const pieceForList = pieceWithSpots.omit({
  description: true,
  recordingLink: true,
  practiceNotes: true,
});

export const pieceFormData = pieceWithSpots
  .omit({ id: true, spots: true })
  .extend({
    id: z.string().optional(),
    spots: z.array(spotFormData),
  });

export const basicPiece = pieceWithSpots.omit({ spots: true });

export const updatePieceWithSpots = pieceFormData.extend({
  id: z.string().optional(),
});

export type PieceForList = z.infer<typeof pieceForList>;
export type BasicPiece = z.infer<typeof basicPiece>;
export type PieceWithSpots = z.infer<typeof pieceWithSpots>;
export type UpdatePieceData = z.infer<typeof updatePieceWithSpots>;
export type PieceFormData = z.infer<typeof pieceFormData>;

export type BasicSpot = z.infer<typeof basicSpot>;
export type SpotFormData = z.infer<typeof spotFormData>;
export type SpotWithPieceInfo = z.infer<typeof spotWithPieceInfo>;
export type UrlOrEmptyForm = z.infer<typeof urlOrEmptyForm>;
export type NotesForm = z.infer<typeof notesForm>;
export type TextForm = z.infer<typeof textForm>;
export type SpotStage = z.infer<typeof spotStage>;
