import { Suspense, lazy } from "preact/compat";

const SpotChart = lazy(() => import("./spot-chart"));

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
    <Suspense fallback={<div>Loading spot breakdown...</div>}>
      <SpotChart
        repeat={parseInt(repeat) || 0}
        extrarepeat={parseInt(extrarepeat) || 0}
        random={parseInt(random) || 0}
        interleave={parseInt(interleave) || 0}
        infrequent={parseInt(infrequent) || 0}
        completed={parseInt(completed) || 0}
      />
    </Suspense>
  );
}
