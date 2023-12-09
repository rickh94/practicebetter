import { type ClassValue, clsx } from "clsx";

export type BasicSpot = {
  id: string;
  name: string;
  measures?: string | null;
};

export type PracticeSummaryItem = {
  name: string;
  reps: number;
  id: string;
};

export type RandomMode = "setup" | "practice" | "summary";

export function uniqueID() {
  return `${Math.floor(Math.random() * Math.random() * Date.now())}`;
}

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

export function getStageDisplayName(stage: string) {
  switch (stage) {
    case "repeat":
      return "Repeat Practice";
    case "more_repeat":
      return "Extra Repeat Practice";
    case "random":
      return "Random Practice";
    case "interleave":
      return "Interleaved Practice";
    case "interleave_days":
      return "Interleave Between Days";
    case "completed":
      return "Completed";
    default:
      return "Unknown";
  }
}
