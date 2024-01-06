import register from "preact-custom-element";
import { PracticeMenu } from "./ui/practice-menu";

try {
  register(PracticeMenu, "practice-menu", ["pieceid"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
