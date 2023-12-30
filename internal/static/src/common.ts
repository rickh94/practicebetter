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
  excellent?: number;
  fine?: number;
  poor?: number;
  day?: number;
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
    case "extra_repeat":
      return "Extra Repeat Practice";
    case "random":
      return "Random Practice";
    case "interleave":
      return "Interleaved Practice";
    case "interleave_days":
      return "Infrequent";
    case "completed":
      return "Completed";
    default:
      return "Unknown";
  }
}

export function getPieceStageDisplayName(stage: string) {
  switch (stage) {
    case "active":
      return "Active";
    case "future":
      return "Not Started";
    case "completed":
      return "Completed";
    default:
      return "Unknown";
  }
}
export type NavItem = {
  label: string;
  href: string;
  icon: preact.JSX.Element;
  highlight?: boolean;
};
