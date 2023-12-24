import register from "preact-custom-element";
import {
  AudioPromptSummary,
  ImagePromptSummary,
  NotesPromptSummary,
} from "./ui/prompts";

try {
  register(AudioPromptSummary, "audio-prompt-summary", ["url"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(ImagePromptSummary, "image-prompt-summary", ["url"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(NotesPromptSummary, "notes-prompt-summary", ["notes"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
