import register from "preact-custom-element";
import { AddSpotsFromPDF } from "./pieces/add-spots-from-pdf";

try {
  register(AddSpotsFromPDF, "add-spots-from-pdf", ["csrf", "pieceid"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
