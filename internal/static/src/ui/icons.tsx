import { cn } from "../common";

export function ShuffleIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--shuffle]")} />;
}

export function RepeatIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--repeat]")} />;
}

export function RandomBoxesIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--random-boxes]")} />;
}

export function NoteSheetIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--note-sheet]")} />;
}

export function PlayListIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--play-list]")} />;
}
