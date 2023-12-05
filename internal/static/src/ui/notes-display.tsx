import ABCJS from "abcjs";
import { useEffect, useRef } from "react";

type NotesDisplayProps = {
  notes: string;
  wrap?: {
    minSpacing: number;
    maxSpacing: number;
    preferredMeasuresPerLine: number;
  };
  staffwidth?: number;
  responsive?: "resize";
};

export default function NotesDisplay({
  notes,
  wrap = undefined,
  staffwidth = undefined,
  responsive = undefined,
}: NotesDisplayProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!ref.current) return;
    ABCJS.renderAbc(ref.current, notes, {
      scale: 1.1,
      add_classes: true,
      paddingleft: 0,
      paddingright: 0,
      paddingbottom: 0,
      paddingtop: 0,
      wrap,
      staffwidth,
      responsive,
    });
  }, [ref, notes, wrap, staffwidth, responsive]);

  return <div className="notes -pl-2 overflow-x-scroll" ref={ref}></div>;
}
