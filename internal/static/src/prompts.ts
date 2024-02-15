import register from "preact-custom-element";
import {
  AudioPromptSummary,
  ImagePromptWC,
  NotesPromptSummary,
} from "./ui/prompts";

try {
  register(AudioPromptSummary, "audio-prompt-summary", [], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(ImagePromptWC, "image-prompt-summary", [], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(NotesPromptSummary, "notes-prompt-summary", [], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
