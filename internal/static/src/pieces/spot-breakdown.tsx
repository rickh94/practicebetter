// import { Suspense, lazy } from "preact/compat";
import SpotChart from "./spot-chart";

// const SpotChart = lazy(() => import("./spot-chart"));

export function SpotBreakdown({
  repeat,
  extrarepeat,
  random,
  interleave,
  infrequent,
  completed,
}: {
  repeat: string;
  extrarepeat: string;
  random: string;
  interleave: string;
  infrequent: string;
  completed: string;
}) {
  return (
    <SpotChart
      repeat={parseInt(repeat, 10) || 0}
      extrarepeat={parseInt(extrarepeat, 10) || 0}
      random={parseInt(random, 10) || 0}
      interleave={parseInt(interleave, 10) || 0}
      infrequent={parseInt(infrequent, 10) || 0}
      completed={parseInt(completed, 10) || 0}
    />
  );
}
