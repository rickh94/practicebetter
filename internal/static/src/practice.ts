import register from "preact-custom-element";
import { RandomSpots } from "./practice/random-spots";
import { SequenceSpots } from "./practice/sequence-spots";
import { Repeat } from "./practice/repeat";
import { StartingPoint } from "./practice/starting-point";

try {
  register(RandomSpots, "random-spots", ["initialspots", "pieceid"], {
    shadow: false,
  });
  register(SequenceSpots, "sequence-spots", ["initialspots", "pieceid"], {
    shadow: false,
  });
  register(Repeat, "repeat-practice", [], { shadow: false });
  register(
    StartingPoint,
    "starting-point",
    ["initialmeasures", "initialbeats", "preconfigured", "pieceid", "csrf"],
    { shadow: false },
  );
} catch (err) {
  console.log(err);
}

async function handleFinishedSpotPracticingEvent(e: CustomEvent) {
  const { spotIDs, durationMinutes, pieceid, csrf } = e.detail;
  if (!spotIDs || spotIDs.length === 0 || !durationMinutes || !pieceid) {
    return;
  }
  const res = await fetch(`/library/pieces/${pieceid}/practice/random-single`, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ spotIDs, durationMinutes }),
  });
  if (res.ok) {
    document.dispatchEvent(
      new CustomEvent("ShowAlert", {
        detail: {
          variant: "success",
          message: "Great job praciticing, keep it up!",
          title: "Practicing Completed",
          duration: 3000,
        },
      }),
    );
  } else {
    console.error(await res.text());
    document.dispatchEvent(
      new CustomEvent("ShowAlert", {
        detail: {
          variant: "error",
          message: "Your practice session could not be saved.",
          title: "Something went wrong",
          duration: 3000,
        },
      }),
    );
  }
}

document.addEventListener(
  "FinishedSpotPraciticing",
  handleFinishedSpotPracticingEvent,
);

async function handleFinishedStartingPointPracticingEvent(e: CustomEvent) {
  const { measuresPracticed, durationMinutes, pieceid, csrf } = e.detail;
  if (!measuresPracticed || !durationMinutes || !pieceid) {
    return;
  }
  const res = await fetch(
    `/library/pieces/${pieceid}/practice/starting-point`,
    {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ measuresPracticed, durationMinutes }),
    },
  );
  if (res.ok) {
    document.dispatchEvent(
      new CustomEvent("ShowAlert", {
        detail: {
          variant: "success",
          message: "Great job praciticing, keep it up!",
          title: "Practicing Completed",
          duration: 3000,
        },
      }),
    );
  } else {
    console.error(await res.text());
    document.dispatchEvent(
      new CustomEvent("ShowAlert", {
        detail: {
          variant: "error",
          message: "Your practice session could not be saved.",
          title: "Something went wrong",
          duration: 3000,
        },
      }),
    );
  }
}

document.addEventListener(
  "FinishedStartingPointPraciticing",
  handleFinishedStartingPointPracticingEvent,
);
