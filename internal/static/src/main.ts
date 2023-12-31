import register from "preact-custom-element";
import { PracticeToolNav } from "./ui/practice-tool-nav";
import { ConfirmDialog } from "./ui/confirm";
import { uniqueID } from "./common";
import { DateFromNow, PrettyDate, NumberDate } from "./ui/date-display";
import { PieceStage, SpotStage } from "./ui/stages";
import { InternalNav } from "./ui/internal-nav";
import { EditRemindersSummary, RemindersSummary } from "./ui/prompts";
import { BackToPiece } from "./ui/links";
import { PracticeSpotDisplayWrapper } from "./practice/practice-spot-display";

try {
  register(
    PracticeSpotDisplayWrapper,
    "practice-spot-display",
    ["spotjson", "pieceid", "piecetitle"],
    { shadow: false },
  );
} catch (err) {
  console.log(err);
}
try {
  register(PracticeToolNav, "practice-tool-nav", ["activepath"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(BackToPiece, "back-to-piece", ["pieceid"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(
    ConfirmDialog,
    "confirm-dialog",
    ["question", "confirmevent", "cancelevent", "id"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
try {
  register(DateFromNow, "date-from-now", ["epoch"], { shadow: false });
} catch (err) {
  console.log(err);
}
try {
  register(PrettyDate, "pretty-date", ["epoch"], { shadow: false });
} catch (err) {
  console.log(err);
}
try {
  register(NumberDate, "number-date", ["epoch"], { shadow: false });
} catch (err) {
  console.log(err);
}
try {
  register(PieceStage, "piece-stage", ["stage"], { shadow: false });
} catch (err) {
  console.log(err);
}
try {
  register(SpotStage, "spot-stage", ["stage", "icon"], { shadow: false });
} catch (err) {
  console.log(err);
}
try {
  register(InternalNav, "internal-nav", ["activeplanid", "activepath"], {
    shadow: false,
  });
} catch (err) {
  console.log(err);
}
try {
  register(
    RemindersSummary,
    "reminders-summary",
    ["text", "pieceid", "spotid"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
try {
  register(
    EditRemindersSummary,
    "edit-reminders-summary",
    ["text", "pieceid", "spotid", "csrf"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
/**
 * Picks the right icon for an alert
 */
function getIcon(variant: AlertVariant) {
  switch (variant) {
    case "success":
      return `
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6 text-green-500">
          <path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12zm13.36-1.814a.75.75 0 10-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.75-5.25z" clip-rule="evenodd" />
        </svg>
        `;
    case "error":
      return `
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-6 h-6 text-red-500">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
        </svg>
        `;
    case "warning":
      return `
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6 text-yellow-500">
          <path fill-rule="evenodd" d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003zM12 8.25a.75.75 0 01.75.75v3.75a.75.75 0 01-1.5 0V9a.75.75 0 01.75-.75zm0 8.25a.75.75 0 100-1.5.75.75 0 000 1.5z" clip-rule="evenodd" />
        </svg>
        `;
    default:
      return `
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6 text-blue-500">
          <path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12zm8.706-1.442c1.146-.573 2.437.463 2.126 1.706l-.709 2.836.042-.02a.75.75 0 01.67 1.34l-.04.022c-1.147.573-2.438-.463-2.127-1.706l.71-2.836-.042.02a.75.75 0 11-.671-1.34l.041-.022zM12 9a.75.75 0 100-1.5.75.75 0 000 1.5z" clip-rule="evenodd" />
        </svg>
        `;
  }
}
// TODO: maybe make this a component

/**
 * Displays an alert
 */
export function showAlert(
  message: string,
  title: string,
  variant: AlertVariant,
  duration: number,
) {
  const toastId = "toast-" + Math.random().toString(36).substring(2, 15);
  const icon = getIcon(variant);

  const toastHTML = `
    <div class="p-4">
      <div class="flex items-start">
        <div class="flex-shrink-0">${icon}</div>
        <div class="flex-1 pt-0.5 ml-3 w-0">
          <p class="text-sm font-medium">${title}</p>
          <p class="mt-1 text-sm">${message}</p>
        </div>
        <div class="flex flex-shrink-0 ml-4">
          <button
            type="button"
            class="inline-flex text-red-700 hover:text-red-500 focus:ring-2 focus:ring-white focus:ring-offset-2 focus:outline-none"
            onclick="document.dispatchEvent(new CustomEvent('CloseAlert', { detail: { id: '${toastId}' } }))"
          >
            <span class="sr-only">Close</span>
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="size-6">
            <path fill-rule="evenodd" d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm-1.72 6.97a.75.75 0 1 0-1.06 1.06L10.94 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06L12 13.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L13.06 12l1.72-1.72a.75.75 0 1 0-1.06-1.06L12 10.94l-1.72-1.72Z" clip-rule="evenodd" />
          </svg>
          </button>
        </div>
      </div>
    </div>
  `;
  const toastDiv = document.createElement("div");
  toastDiv.className =
    "overflow-hidden w-full max-w-sm bg-white rounded-xl ring-1 ring-black ring-opacity-5 shadow-lg transition-transform pointer-events-auto text-neutral-800 border-neutral-800 shadow-neutral-800/20";
  toastDiv.id = toastId;
  toastDiv.classList.add("transform", "ease-out", "duration-300", "transition");
  toastDiv.classList.add(
    "translate-y-2",
    "opacity-0",
    "sm:translate-y-0",
    "sm:translate-x-2",
  );
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      toastDiv.classList.remove(
        "translate-y-2",
        "opacity-0",
        "sm:translate-y-0",
        "sm:translate-x-2",
      );
      toastDiv.classList.add(
        "translate-y-0",
        "opacity-100",
        "sm:translate-x-0",
        "sm:translate-y-0",
      );
      setTimeout(() => {
        toastDiv.classList.remove(
          "translate-y-0",
          "opacity-100",
          "sm:translate-x-0",
          "sm:translate-y-0",
        );
        toastDiv.classList.remove(
          "transform",
          "ease-out",
          "duration-300",
          "transition",
        );
      }, 300);
    });
  });
  toastDiv.innerHTML = toastHTML;
  const toastContainer = document.getElementById("toast-container");
  if (!toastContainer) {
    return;
  }
  toastContainer.append(toastDiv);
  setTimeout(() => {
    closeAlert(toastId);
  }, duration);
}

export function closeAlert(id: string) {
  const toastDiv = document.getElementById(id);
  if (!toastDiv) return;
  toastDiv.classList.add("transform", "ease-in", "duration-300", "transition");
  toastDiv.classList.add("opacity-100");

  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      toastDiv.classList.remove("opacity-100");
      toastDiv.classList.add("opacity-0");
      setTimeout(() => {
        toastDiv.remove();
      }, 300);
    });
  });
}

function handleCloseAlertEvent(evt: CloseAlertEvent) {
  if (!evt.detail) {
    throw new Error("Invalid event received from server");
  }
  const { id } = evt.detail;
  if (!id) {
    throw new Error("Invalid event received from server");
  }
  closeAlert(id);
}

type CloseAlertEvent = Event & {
  detail?: {
    id?: string;
  };
};

type ShowAlertEvent = Event & {
  detail?: {
    message?: string;
    title?: string;
    variant?: AlertVariant;
    duration?: number;
  };
};

type AlertVariant = "success" | "info" | "warning" | "error";

function handleAlertEvent(evt: ShowAlertEvent) {
  if (!evt.detail) {
    throw new Error("Invalid event received from server");
  }
  const { message, title, variant, duration } = evt.detail;
  if (!message || !title || !variant || !duration) {
    throw new Error("Invalid event received from server");
  }
  showAlert(message, title, variant, duration);
}

type FocusInputEvent = Event & {
  detail?: {
    id?: string;
  };
};

function handleFocusInputEvent(evt: FocusInputEvent) {
  if (!evt.detail) {
    throw new Error("Invalid event received from server");
  }
  const { id } = evt.detail;
  if (!id) {
    throw new Error("Invalid event received from server");
  }
  const foundInput = document.getElementById(id);
  if (!foundInput || !(foundInput instanceof HTMLInputElement)) {
    throw new Error("Invalid event received from server");
  }
  foundInput.focus();
  foundInput.select();
}
// TODO: maybe switch this to a web component

document.addEventListener("ShowAlert", handleAlertEvent);
document.addEventListener("FocusInput", handleFocusInputEvent);
document.addEventListener("CloseAlert", handleCloseAlertEvent);

type HTMXConfirmEvent = Event & {
  detail: {
    question?: string;
    issueRequest?: (response: boolean) => void;
  };
};

function closeModal(id: string) {
  const modal = document.getElementById(id);
  if (!modal) return;
  modal.classList.add("close");

  requestAnimationFrame(function () {
    requestAnimationFrame(function () {
      modal.classList.remove("close");
      // @ts-ignore
      modal.close();
      setTimeout(() => modal.remove(), 1000);
    });
  });
}

document.addEventListener("htmx:confirm", function (e: HTMXConfirmEvent) {
  const { question, issueRequest } = e.detail;
  if (!question) {
    return;
  }
  if (!issueRequest) {
    return;
  }
  e.preventDefault();
  const id = uniqueID();
  const confirmevent = `${id}confirm`;
  const cancelevent = `${id}cancel`;

  function onConfirm() {
    closeModal(id);
    issueRequest(true);
    document.removeEventListener(confirmevent, onConfirm);
    document.removeEventListener(cancelevent, onCancel);
  }
  function onCancel() {
    closeModal(id);
    document.removeEventListener(confirmevent, onConfirm);
    document.removeEventListener(cancelevent, onCancel);
  }

  document.addEventListener(confirmevent, onConfirm);
  document.addEventListener(cancelevent, onCancel);

  const dialog = document.createElement("confirm-dialog");
  dialog.setAttribute("dialogid", id);
  dialog.setAttribute("question", question);
  dialog.setAttribute("confirmevent", confirmevent);
  dialog.setAttribute("cancelevent", cancelevent);

  document.getElementById("main-content").appendChild(dialog);
  (document.getElementById(id) as HTMLDialogElement).showModal();
});

document.addEventListener("htmx:afterSwap", (event: CustomEvent) => {
  if (!(event.target instanceof HTMLElement)) {
    return;
  }
  if (event.detail?.target?.id === "main-content") {
    window.scrollTo(0, 0);
  }
});
