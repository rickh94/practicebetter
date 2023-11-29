import register from "preact-custom-element";
import {
  TextPromptSummary,
  AudioPromptSummary,
  ImagePromptSummary,
  NotesPromptSummary,
} from "./ui/prompts";

try {
  register(TextPromptSummary, "text-prompt-summary", ["text"], {
    shadow: false,
  });
  register(AudioPromptSummary, "audio-prompt-summary", ["url"], {
    shadow: false,
  });
  register(ImagePromptSummary, "image-prompt-summary", ["url"], {
    shadow: false,
  });
  register(NotesPromptSummary, "notes-prompt-summary", ["notes"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
