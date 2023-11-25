import { Line } from "react-chartjs-2";
import {
  Chart as ChartJS,
  ArcElement,
  LinearScale,
  CategoryScale,
  PointElement,
  LineElement,
  Filler,
} from "chart.js";

ChartJS.register(
  ArcElement,
  LinearScale,
  PointElement,
  LineElement,
  Filler,
  CategoryScale,
);

export function Chart() {
  return (
    <Line
      options={{
        responsive: true,
        scales: {
          x: {
            type: "category",
            labels: ["Start of piece", "", "", "", "", "End of piece"],
          },
          y: {
            beginAtZero: true,
            ticks: {
              display: false,
            },
            title: {
              display: true,
              text: "Total Practice Time",
            },
          },
        },
        plugins: {
          legend: {
            display: false,
          },
          tooltip: {
            enabled: false,
          },
        },
      }}
      data={{
        datasets: [
          {
            fill: true,
            label: "Descending line chart",
            data: [16, 8, 4, 2, 1, 0.5],
            backgroundColor: "rgba(249, 168, 212, 0.5)", // tailwind pink 300
            borderColor: "rgba(249, 168, 212, 1)",
          },
        ],
      }}
      className="self-center p-4 max-w-2xl rounded shadow bg-neutral-50/60"
    />
  );
}
