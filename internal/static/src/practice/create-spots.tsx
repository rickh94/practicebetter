import { useAutoAnimate } from "@formkit/auto-animate/preact";
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
                <div>Remove All</div>
                <span className="sr-only">Remove All Spots</span>
                <span
                  className="icon-[iconamoon--sign-minus-circle-thin] -mr-2 size-5"
                  aria-hidden="true"
                />
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
          <li key={spot.id} className="flex h-10 items-center rounded-xl p-0">
            <div className="whitespace-nowrap rounded-l-xl border-neutral-800 bg-neutral-700/10 py-2 pl-3 pr-2">
              {spot.name}
            </div>
            <button
              onClick={() => deleteSpot(spot.id)}
              className="flex h-10 items-center rounded-r-xl border-red-800 bg-red-700/10 px-3 text-red-800 hover:bg-red-500/10 hover:text-red-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-800 focus-visible:ring-offset-1 focus-visible:ring-offset-neutral-100"
            >
              <span className="sr-only">Delete {spot.name}</span>
              <span
                className="icon-[iconamoon--sign-minus-circle-thin] size-5"
                aria-hidden="true"
              />
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
        <div className="flex flex-wrap gap-2">
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
          <BasicButton onClick={onAddSpot} className="flex-shrink-0">
            <span
              className="icon-[iconamoon--sign-plus-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
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
            <span
              className="icon-[iconamoon--sign-plus-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Add Spots
          </BasicButton>
        </div>
      </div>
    </div>
  );
}
