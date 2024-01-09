import { cn } from "../common";

export function ShuffleIcon({ className }: { className?: string }) {
  return (
    <span
      className={cn(className, "icon-[iconamoon--playlist-shuffle-thin]")}
    />
  );
}

export function RepeatIcon({ className }: { className?: string }) {
  return (
    <span
      className={cn(className, "icon-[iconamoon--playlist-repeat-list-thin]")}
    />
  );
}

export function RandomBoxesIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--random-boxes]")} />;
}

export function MusicFileIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--music-file]")} />;
}

export function PlayListIcon({ className }: { className?: string }) {
  return <span className={cn(className, "icon-[custom--play-list]")} />;
}
