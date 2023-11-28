import register from "preact-custom-element";
import { EditPieceForm } from "./pieces/edit";

try {
  register(
    EditPieceForm,
    "edit-piece-form",
    ["csrf", "piece", "spots", "pieceid"],
    { shadow: false },
  );
} catch (err) {
  console.log(err);
}
