import { type ClassValue, clsx } from "clsx";
import { type JSX } from "preact/jsx-runtime";

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
  icon: JSX.Element;
  highlight?: boolean;
};

export type CroppedImageData = {
  data: string;
  id: string;
  width: number;
  height: number;
  x?: number;
  y?: number;
  transformationMatrix?: number[];
};

export type PageImage = {
  src: string;
  alt: string;
  id: string;
};
