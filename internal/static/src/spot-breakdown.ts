import register from "preact-custom-element";
import { SpotBreakdown } from "./pieces/spot-breakdown";

try {
  register(
    SpotBreakdown,
    "spot-breakdown",
    [
      "repeat",
      "extrarepeat",
      "random",
      "interleave",
      "infrequent",
      "completed",
    ],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
