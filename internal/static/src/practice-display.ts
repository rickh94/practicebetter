import register from "preact-custom-element";
import { PastPracticeDisplay } from "./past-practice/display";

try {
  register(
    PastPracticeDisplay,
    "past-practice-display",
    ["sessions", "title", "wide", "background"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
