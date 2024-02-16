import { useAutoAnimate } from "@formkit/auto-animate/preact";
import { uniqueID } from "../common";
import { useCallback, useRef, type StateUpdater } from "preact/hooks";
import { type BasicSpot } from "../validators";

export function CreateSpots({
  setSpots,
  spots,
}: {
  setSpots: StateUpdater<BasicSpot[]>;
  spots: BasicSpot[];
}) {
  const spotNameRef = useRef<HTMLInputElement>(null);
  const numSpotsRef = useRef<HTMLInputElement>(null);
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  const [parent] = useAutoAnimate<HTMLUListElement>();

  const onAddSpot = useCallback(() => {
    if (!spotNameRef.current) return;
    const name = spotNameRef.current.value;
    setSpots((prev) => [
      ...prev,
      { name, id: uniqueID(), stage: "random", measures: "" },
    ]);
    spotNameRef.current.value = "";
  }, [setSpots, spotNameRef]);

  const generateSomeSpots = useCallback(() => {
    if (!numSpotsRef.current) return;
    const numSpots = parseInt(numSpotsRef.current?.value, 10);
    const tmpSpots: BasicSpot[] = [];
    for (let i = spots.length; i < numSpots + spots.length; i++) {
      tmpSpots.push({
        name: `Spot #${i + 1}`,
        id: uniqueID(),
        stage: "random",
        measures: "",
      });
    }
    setSpots((prev) => [...prev, ...tmpSpots]);
  }, [setSpots, spots, numSpotsRef]);

  const deleteSpot = useCallback(
    (spotId: string) => {
      setSpots((prev) => prev.filter((spot) => spot.id !== spotId));
    },
    [setSpots],
  );

  const clearSpots = useCallback(() => {
    setSpots([]);
  }, [setSpots]);

  return (
    <div className="grid w-full grid-cols-1 gap-2 md:grid-cols-2">
      <div className="col-span-full flex-col">
        <div className="flex flex-row items-center gap-4">
          <div>
            <h2 className="text-xl font-bold">Your Spots</h2>
          </div>
          <div>
            {spots.length > 0 && (
              <button
                className="action-button red focusable"
                type="button"
                onClick={() => clearSpots()}
              >
                <div>Remove All</div>
                <span className="sr-only">Remove All Spots</span>
                <span
                  className="icon-[iconamoon--sign-minus-circle-thin] -mr-2 size-5"
                  aria-hidden="true"
                />
              </button>
            )}
          </div>
        </div>
        {spots.length === 0 && (
          <p className="text-sm text-neutral-700">Add some spots below</p>
        )}
      </div>
      <ul
        className="col-span-full flex w-full flex-wrap gap-3"
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        ref={parent}
      >
        {spots.map((spot) => (
          <li
            key={spot.id}
            className="flex h-10 items-center rounded-xl p-0 shadow-sm shadow-neutral-900/30"
          >
            <div className="flex h-10 items-center whitespace-nowrap rounded-l-xl border-y border-l border-neutral-400 bg-neutral-200 pl-3 pr-3 shadow-sm">
              {spot.name}
            </div>
            <button
              onClick={() => spot.id && deleteSpot(spot.id)}
              className="flex h-10 items-center rounded-r-xl border-y border-r border-red-400 bg-red-200 px-3 text-red-800 hover:border-red-500 hover:bg-red-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-800 focus-visible:ring-offset-1 focus-visible:ring-offset-neutral-100"
            >
              <span className="sr-only">Delete {spot.name}</span>
              <span
                className="icon-[iconamoon--sign-minus-circle-thin] size-6"
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
            className="basic-field w-44"
            placeholder="Spot #1"
            type="text"
            onKeyUp={(e) => {
              if (e.key === "Enter" || e.key === ",") {
                onAddSpot();
              }
            }}
          />
          <button
            className="action-button neutral focusable"
            type="button"
            onClick={onAddSpot}
          >
            <span
              className="icon-[iconamoon--sign-plus-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Add Spot
          </button>
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
            className="basic-field w-20"
            placeholder="10"
            type="number"
            min="1"
            ref={numSpotsRef}
            onKeyUp={(e) => {
              if (e.key === "Enter") {
                generateSomeSpots();
              }
            }}
          />
          <button
            type="button"
            className="action-button neutral focusable"
            onClick={generateSomeSpots}
          >
            <span
              className="icon-[iconamoon--sign-plus-circle-thin] -ml-1 size-6"
              aria-hidden="true"
            />
            Add Spots
          </button>
        </div>
      </div>
    </div>
  );
}
