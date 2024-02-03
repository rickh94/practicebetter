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
    ["initialspot", "pieceid", "csrf", "piecetitle", "planid", "kidmode"],
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

globalThis.addEventListener("FinishedSpotPracticing", (e) => {
  const { spots, durationMinutes, csrf, endpoint } = e.detail;
  if (!spots || spots.length === 0 || !durationMinutes || !csrf || !endpoint) {
    console.log(e.detail);
    console.error("event missing data");
    return;
  }
  fetch(endpoint, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ spots, durationMinutes }),
  })
    .then((res) => {
      console.log(res);
      if (res.ok) {
        globalThis.dispatchEvent(
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
        res.text().then(console.error).catch(console.error);
        globalThis.dispatchEvent(
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
    })
    .catch((err) => {
      console.error(err);
      globalThis.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            variant: "error",
            message: "Your practice session could not be saved.",
            title: "Something went wrong",
            duration: 3000,
          },
        }),
      );
    });
});

globalThis.addEventListener("FinishedStartingPointPracticing", (evt) => {
  const { measuresPracticed, durationMinutes, pieceid, csrf } = evt.detail;
  if (!measuresPracticed || !durationMinutes || !pieceid) {
    console.error("event missing data");
    return;
  }
  fetch(`/library/pieces/${pieceid}/practice/starting-point`, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ measuresPracticed, durationMinutes }),
  })
    .then((res) => {
      if (res.ok) {
        globalThis.dispatchEvent(
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
        res.text().then(console.error).catch(console.error);
        globalThis.dispatchEvent(
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
    })
    .catch((err) => {
      console.error(err);
      globalThis.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            variant: "error",
            message: "Your practice session could not be saved.",
            title: "Something went wrong",
            duration: 3000,
          },
        }),
      );
    });
});

globalThis.addEventListener("FinishedRepeatPracticing", (evt) => {
  const { durationMinutes, csrf, endpoint, success, toStage } = evt.detail;
  if (!durationMinutes || !csrf || !endpoint) {
    console.error("event missing data");
    return;
  }
  fetch(endpoint, {
    method: "POST",
    headers: {
      "X-CSRF-Token": csrf,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ durationMinutes, success, toStage }),
  })
    .then((res) => {
      if (res.ok) {
        if (success) {
          globalThis.dispatchEvent(
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
          globalThis.dispatchEvent(
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
        res.text().then(console.error).catch(console.error);
        globalThis.dispatchEvent(
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
    })
    .catch((err) => {
      console.error(err);
      globalThis.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            variant: "error",
            message: "Your practice session could not be saved.",
            title: "Something went wrong",
            duration: 3000,
          },
        }),
      );
    });
});
