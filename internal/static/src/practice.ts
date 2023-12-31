import register from "preact-custom-element";
import { RandomSpots } from "./practice/random-spots";
import { Repeat } from "./practice/repeat";
import { StartingPoint } from "./practice/starting-point";

try {
  register(
    RandomSpots,
    "random-spots",
    ["initialspots", "pieceid", "csrf", "initialsessions", "planid"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
try {
  register(
    Repeat,
    "repeat-practice",
    ["initialspot", "pieceid", "csrf", "piecetitle", "planid"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
try {
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
  const { spots, durationMinutes, csrf, endpoint } = e.detail;
  if (!spots || spots.length === 0 || !durationMinutes || !csrf || !endpoint) {
    console.log(e.detail);
    console.error("event missing data");
    return;
  }
  const res = await fetch(endpoint, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ spots, durationMinutes }),
  });
  if (res.ok) {
    document.dispatchEvent(
      new CustomEvent("ShowAlert", {
        detail: {
          variant: "success",
          message: "Great job practicing, keep it up!",
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
  "FinishedSpotPracticing",
  handleFinishedSpotPracticingEvent,
);

async function handleFinishedStartingPointPracticingEvent(e: CustomEvent) {
  const { measuresPracticed, durationMinutes, pieceid, csrf } = e.detail;
  if (!measuresPracticed || !durationMinutes || !pieceid) {
    console.error("event missing data");
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
          message: "Great job practicing, keep it up!",
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
  "FinishedStartingPointPracticing",
  handleFinishedStartingPointPracticingEvent,
);

async function handleFinishedRepeatPracticingEvent(e: CustomEvent) {
  const { durationMinutes, csrf, endpoint, success, toStage } = e.detail;
  if (!durationMinutes || !csrf || !endpoint) {
    console.error("event missing data");
    return;
  }
  const res = await fetch(endpoint, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ durationMinutes, success, toStage }),
  });
  if (res.ok) {
    if (success) {
      document.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            variant: "success",
            message:
              "Great job practicing, you can now start to randomly practice this spot!",
            title: "Practicing Completed",
            duration: 3000,
          },
        }),
      );
    } else {
      document.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            variant: "info",
            message:
              "Great job practicing, come back again to get it five times in a row!",
            title: "Practicing Completed",
            duration: 3000,
          },
        }),
      );
    }
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
  "FinishedRepeatPracticing",
  handleFinishedRepeatPracticingEvent,
);
