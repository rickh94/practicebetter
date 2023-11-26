import register from "preact-custom-element";
import { CreatePieceForm } from "./pieces/create";

try {
  register(CreatePieceForm, "create-piece-form", ["csrf"], { shadow: false });
} catch (err) {
  console.log(err);
}
