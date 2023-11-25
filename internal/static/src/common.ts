import { twMerge } from "tailwind-merge";
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
  return twMerge(clsx(inputs));
}
