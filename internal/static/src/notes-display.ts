import register from "preact-custom-element";
import NotesDisplay from "./ui/notes-display";

try {
  register(
    NotesDisplay,
    "notes-display",
    ["notes", "wrap", "staffWidth", "responsive"],
    { shadow: false },
  );
} catch (err) {
  console.log(err);
}
