import { Doughnut } from "react-chartjs-2";
import {
  Chart as ChartJS,
  ArcElement,
  CategoryScale,
  PointElement,
  Filler,
  Tooltip,
  Legend,
} from "chart.js";
import { ChartPieIcon, ChevronRightIcon } from "@heroicons/react/24/solid";

ChartJS.register(
  ArcElement,
  PointElement,
  Filler,
  CategoryScale,
  Tooltip,
  Legend,
);

export default function SpotChart({
  repeat,
  extrarepeat,
  random,
  interleave,
  infrequent,
  completed,
}: {
  repeat: number;
  extrarepeat: number;
  random: number;
  interleave: number;
  infrequent: number;
  completed: number;
}) {
  const options = {
    responsive: true,
    plugins: {
      legend: {
        display: true,
      },
      tooltip: {
        enabled: true,
      },
    },
  };
  const data = {
    labels: [
      "Repeat",
      "Extra Repeat",
      "Random",
      "Interleave",
      "Infrequent",
      "Completed",
    ],
    datasets: [
      {
        label: "Spots",
        data: [repeat, extrarepeat, random, interleave, infrequent, completed],
        backgroundColor: [
          "#fde68a", // tailwind amber 200
          "#fed7aa", // tailwind amber 200
          "#fbcfe8", // tailwind pink 200
          "#c7d2fe", // tailwind indigo 200
          "#bae6fd", // tailwind sky 200
          "#bbf7d0", // tailwind green 200
        ],
        borderColor: "#737373",
        borderWidth: 1,
      },
    ],
  };
  return (
    <>
      <div className="hidden w-full flex-col rounded-xl bg-neutral-700/5 p-4 md:flex">
        <h4 className="text-center text-xl font-bold">Spots Progress Chart</h4>
        <Doughnut options={options} data={data} />
      </div>
      <details className="flex w-full flex-col rounded-xl bg-neutral-700/5 p-4 md:hidden">
        <summary className="flex w-full cursor-pointer items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <ChartPieIcon className="size-5" />
            <h4 className="text-center text-xl font-bold">
              Spots Progress Chart
            </h4>
          </div>
          <ChevronRightIcon className="summary-icon size-6 transition-transform" />
        </summary>
        <div className="mx-auto max-w-md">
          <Doughnut options={options} data={data} />
        </div>
      </details>
    </>
  );
}
