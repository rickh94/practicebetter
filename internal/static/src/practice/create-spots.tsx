import { useAutoAnimate } from "@formkit/auto-animate/preact";
import { PlusCircleIcon, TrashIcon } from "@heroicons/react/20/solid";
import { uniqueID } from "../common";
import { AngryButton, BasicButton } from "../ui/buttons";
import { useCallback, useRef } from "preact/compat";
import { StateUpdater } from "preact/hooks";
import { BasicSpot } from "../validators";

export function CreateSpots({
  setSpots,
  spots,
}: {
  setSpots: StateUpdater<BasicSpot[]>;
  spots: BasicSpot[];
}) {
  const spotNameRef = useRef<HTMLInputElement>(null);
  const numSpotsRef = useRef<HTMLInputElement>(null);
  const [parent] = useAutoAnimate();

  const onAddSpot = useCallback(
    function () {
      if (!spotNameRef.current) return;
      const name = spotNameRef.current.value;
      setSpots((prev) => [...prev, { name, id: uniqueID(), stage: "random" }]);
      spotNameRef.current.value = "";
    },
    [setSpots, spotNameRef],
  );

  const generateSomeSpots = useCallback(
    function () {
      if (!numSpotsRef.current) return;
      const numSpots = parseInt(numSpotsRef.current?.value);
      const tmpSpots: BasicSpot[] = [];
      for (let i = spots.length; i < numSpots + spots.length; i++) {
        tmpSpots.push({
          name: `Spot #${i + 1}`,
          id: uniqueID(),
        });
      }
      setSpots((prev) => [...prev, ...tmpSpots]);
    },
    [setSpots, spots, numSpotsRef],
  );

  const deleteSpot = useCallback(
    function (spotId: string) {
      setSpots((prev) => prev.filter((spot) => spot.id !== spotId));
    },
    [setSpots],
  );

  const clearSpots = useCallback(
    function () {
      setSpots([]);
    },
    [setSpots],
  );

  return (
    <div className="grid w-full grid-cols-1 gap-2 md:grid-cols-2">
      <div className="col-span-full flex-col">
        <div className="flex flex-row items-center gap-4">
          <div>
            <h2 className="text-xl font-bold">Your Spots</h2>
          </div>
          <div>
            {spots.length > 0 && (
              <AngryButton onClick={() => clearSpots()}>
                <div>Delete All</div>
                <span className="sr-only">Delete All Spots</span>
                <TrashIcon className="h-4 w-4" />
              </AngryButton>
            )}
          </div>
        </div>
        {spots.length === 0 && (
          <p className="text-sm text-neutral-700">Add some spots below</p>
        )}
      </div>
      <ul className="col-span-full flex w-full flex-wrap gap-3" ref={parent}>
        {spots.map((spot) => (
          <li key={spot.id} className="flex items-center rounded-xl p-0">
            <div className="whitespace-nowrap rounded-l-xl border-neutral-800 bg-neutral-700/10 py-2 pl-3 pr-2">
              {spot.name}
            </div>
            <button
              onClick={() => deleteSpot(spot.id)}
              className="rounded-r-xl border-red-800 bg-red-700/10 py-2.5 pl-2 pr-3 text-red-800 hover:bg-red-500/10 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-800 focus:ring-offset-1 focus:ring-offset-neutral-100"
            >
              <span className="sr-only">Delete {spot.name}</span>
              <TrashIcon className="h-5 w-5" />
            </button>
          </li>
        ))}
      </ul>
      <div className="flex flex-col">
        <label
          className="text-lg font-semibold text-neutral-800"
          htmlFor="spot-name"
        >
          Spot Name
        </label>
        <p className="pb-2 text-sm text-neutral-700">
          Enter the name of a spot
        </p>
        <div className="flex gap-2">
          <input
            id="spot-name"
            ref={spotNameRef}
            className="focusable w-44 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20"
            type="text"
            onKeyUp={(e) => {
              if (e.key === "Enter" || e.key === ",") {
                onAddSpot();
              }
            }}
          />
          <BasicButton onClick={onAddSpot}>
            <PlusCircleIcon className="-ml-1 h-4 w-4" />
            Add Spot
          </BasicButton>
        </div>
      </div>
      <div className="flex flex-col">
        <label
          className="text-lg font-semibold text-neutral-800"
          for="num-spots"
        >
          Generate Spots
        </label>
        <p className="pb-2 text-sm text-neutral-700">
          Automatically add a number of spots.
        </p>
        <div className="flex gap-2">
          <input
            id="num-spots"
            className="focusable w-20 rounded-xl bg-neutral-700/10 px-4 py-2 font-semibold text-neutral-800 transition duration-200 focus:bg-neutral-700/20"
            type="number"
            min="1"
            ref={numSpotsRef}
            onKeyUp={(e) => {
              if (e.key === "Enter") {
                generateSomeSpots();
              }
            }}
          />
          <BasicButton onClick={generateSomeSpots}>
            <PlusCircleIcon className="-ml-1 h-4 w-4" />
            Add Spots
          </BasicButton>
        </div>
      </div>
    </div>
  );
}
