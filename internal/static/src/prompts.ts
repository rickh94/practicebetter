import register from "preact-custom-element";
import {
  RemindersSummary,
  AudioPromptSummary,
  ImagePromptSummary,
  NotesPromptSummary,
  EditRemindersSummary,
} from "./ui/prompts";

try {
  register(
    RemindersSummary,
    "reminders-summary",
    ["text", "pieceid", "spotid"],
    {
      shadow: false,
    },
  );
  register(
    EditRemindersSummary,
    "edit-reminders-summary",
    ["text", "pieceid", "spotid", "csrf"],
    {
      shadow: false,
    },
  );
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
