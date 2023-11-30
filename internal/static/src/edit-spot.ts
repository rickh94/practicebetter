import register from "preact-custom-element";
import { EditSpotForm } from "./pieces/edit-spot";

try {
  register(
    EditSpotForm,
    "edit-spot-form",
    ["csrf", "pieceid", "spotdata", "spotid"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
