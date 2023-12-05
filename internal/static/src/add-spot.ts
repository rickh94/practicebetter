import register from "preact-custom-element";
import { AddSpotForm } from "./pieces/add-spot";

try {
  register(
    AddSpotForm,
    "add-spot-form",
    ["csrf", "pieceid", "initialspotcount"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
