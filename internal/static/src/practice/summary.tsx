import { Cog6ToothIcon, MusicalNoteIcon } from "@heroicons/react/24/solid";
import { type PracticeSummaryItem } from "../common";
import { HappyButton, WarningButton } from "../ui/buttons";
import { BackToPieceLink } from "../ui/links";
// import { BackToPieceLink } from "@ui/links";

// TODO: implement some kind of sorting, or find a way to retain entered order,
// possibly switch to a hashmap
export default function Summary({
  summary,
  setup,
  practice,
  pieceHref,
}: {
  summary: PracticeSummaryItem[];
  setup: () => void;
  practice: () => void;
  pieceHref?: string;
}) {
  return (
    <>
      <div className="flex w-full flex-col justify-center gap-x-8 gap-y-2 px-4 pt-12 sm:flex-row sm:gap-x-6 sm:px-0">
        {pieceHref && <BackToPieceLink pieceHref={pieceHref} />}
        <WarningButton onClick={setup}>
          <Cog6ToothIcon className="-ml-1 h-5 w-5" />
          Back to Setup
        </WarningButton>
        <HappyButton onClick={practice}>
          <MusicalNoteIcon className="-ml-1 h-5 w-5" />
          Practice More
        </HappyButton>
      </div>
      <h2 className="w-full pt-12 text-center text-2xl font-semibold">
        Practice Summary
      </h2>
      <div className="flex w-full flex-col items-center justify-center gap-2 pt-4">
        <table className="min-w-full divide-y divide-neutral-700">
          <thead>
            <tr>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:pl-0"
              >
                Spot Name
              </th>
              <th
                scope="col"
                className="py-3 pl-4 pr-3 text-center text-xs font-medium uppercase tracking-wide text-neutral-500 sm:pl-0"
              >
                Times Practiced
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-neutral-700">
            {summary.map(({ name, reps, id }, idx) => (
              <tr
                key={id}
                className={`${idx % 2 === 0 && "bg-neutral-700/10"}`}
              >
                <td className="whitespace-nowrap py-2 pl-4 pr-3 text-center font-medium text-neutral-900 sm:pl-0">
                  {name}
                </td>
                <td className="whitespace-nowrap px-3 py-2 text-center text-neutral-800">
                  {reps}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
