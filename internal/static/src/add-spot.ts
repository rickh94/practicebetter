import register from "preact-custom-element";
import { AddSpotForm } from "./pieces/add-spot";

try {
  register(AddSpotForm, "add-spot-form", ["csrf", "pieceid"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
