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
import "./input.css";
import "htmx.org";
import * as SimpleWebAuthnBrowser from "@simplewebauthn/browser";
import {
  PublicKeyCredentialRequestOptionsJSON,
  PublicKeyCredentialCreationOptionsJSON,
} from "@simplewebauthn/typescript-types";

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
  register(InternalNav, "internal-nav", ["activepath"], {
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
      return `<span class="text-green-500 size-6 icon-[iconamoon--check-circle-1-thin]"></span>`;
    case "error":
      return `<span class="text-red-500 size-6 icon-[iconamoon--shield-no-thin]"></span>`;
    case "warning":
      return `<span class="text-yellow-500 size-6 icon-[ph--warning-thin]"></span>`;
    default:
      return `<span class="text-blue-500 size-6 icon-[iconamoon--information-circle-thin]"></span>`;
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
            class="inline-flex hover:text-red-800 focus:ring-2 focus:ring-white focus:ring-offset-2 focus:outline-none text-red-800/50"
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
  toastDiv.classList.add("-translate-y-2", "opacity-0");
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      toastDiv.classList.remove("-translate-y-2", "opacity-0");
      toastDiv.classList.add("translate-y-0", "opacity-100");
      setTimeout(() => {
        toastDiv.classList.remove("translate-y-0", "opacity-100");
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

  toastDiv.remove();
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

globalThis.startPasskeyAuth = function (
  publicKey: PublicKeyCredentialRequestOptionsJSON,
  csrf: string,
  nextLoc: string,
) {
  SimpleWebAuthnBrowser.startAuthentication(publicKey)
    .then((attResp) => {
      fetch("/auth/passkey/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
        body: JSON.stringify(attResp),
      })
        .then((res) => res.json())
        .then((res) => {
          if (res.status == "ok") {
            window.location.href = nextLoc;
          } else {
            console.log(res);
            document.dispatchEvent(
              new CustomEvent("ShowAlert", {
                detail: {
                  message: "Could not login",
                  title: "Error",
                  variant: "error",
                  duration: 3000,
                },
              }),
            );
          }
        })
        .catch((err) => {
          console.log(err);
          document.dispatchEvent(
            new CustomEvent("ShowAlert", {
              detail: {
                message: "Could not login",
                title: "Error",
                variant: "error",
                duration: 3000,
              },
            }),
          );
        });
    })
    .catch((err) => {
      console.log(err);
      document.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            message: "Could not login",
            title: "Error",
            variant: "error",
            duration: 3000,
          },
        }),
      );
    });
};

globalThis.startPasskeyRegistration = function (
  publicKey: PublicKeyCredentialCreationOptionsJSON,
  csrf: string,
) {
  SimpleWebAuthnBrowser.startRegistration(publicKey)
    .then((attResp) => {
      fetch("/auth/passkey/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
        body: JSON.stringify(attResp),
      })
        .then((res) => {
          console.log(res);
          if (res.ok) {
            document.dispatchEvent(
              new CustomEvent("ShowAlert", {
                detail: {
                  message: "Use your passkey to login in the future!",
                  title: "Passkey Registered",
                  variant: "success",
                  duration: 3000,
                },
              }),
            );
            document.getElementById("passkey-count").innerHTML = (
              parseInt(document.getElementById("passkey-count").innerHTML) + 1
            ).toString();
          } else {
            console.log(res);
            document.dispatchEvent(
              new CustomEvent("ShowAlert", {
                detail: {
                  message: "Could not register your new passkey",
                  title: "Getistration Error",
                  variant: "error",
                  duration: 3000,
                },
              }),
            );
          }
        })
        .catch((err) => {
          console.log(err);
          document.dispatchEvent(
            new CustomEvent("ShowAlert", {
              detail: {
                message: "Could not regsiter your passkey",
                title: "Error",
                variant: "error",
                duration: 3000,
              },
            }),
          );
        });
    })
    .catch((err) => {
      console.log(err);
      document.dispatchEvent(
        new CustomEvent("ShowAlert", {
          detail: {
            message: "Could not register your passkey",
            title: "Error",
            variant: "error",
            duration: 3000,
          },
        }),
      );
    });
};
