import register from "preact-custom-element";
import { PracticeToolNav } from "./ui/practice-tool-nav";
import { ConfirmDialog } from "./ui/confirm";
import { uniqueID } from "./common";
import { DateFromNow, PrettyDate, NumberDate } from "./ui/date-display";
import { PieceStage, SpotStage } from "./ui/stages";
import { InternalNav } from "./ui/internal-nav";
import { RemindersSummary } from "./ui/prompts";
import { BackToPiece } from "./ui/links";
import { PracticeSpotDisplayWrapper } from "./practice/practice-spot-display";
import "./input.css";
import "htmx.org";
import * as SimpleWebAuthnBrowser from "@simplewebauthn/browser";
import type {
  PublicKeyCredentialRequestOptionsJSON,
  PublicKeyCredentialCreationOptionsJSON,
} from "@simplewebauthn/typescript-types";
import type {
  AlertVariant,
  CloseAlertEvent,
  CloseModalEvent,
  FocusInputEvent,
  HTMXConfirmEvent,
  HTMXRequestEvent,
  ShowAlertEvent,
} from "./types";
import { BreakDialog } from "./ui/plan-components";

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
    ["text", "csrf", "pieceid", "spotid"],
    {
      shadow: false,
    },
  );
} catch (err) {
  console.log(err);
}
try {
  register(BreakDialog, "break-dialog", ["open", "csrf"], { shadow: false });
} catch (err) {
  console.log(err);
}
/**
 * Picks the right icon for an alert
 */
function getIcon(variant: AlertVariant) {
  switch (variant) {
    case "success":
      return `<span class="text-green-500 size-6 icon-[iconamoon--check-circle-1-fill]"></span>`;
    case "error":
      return `<span class="text-red-500 size-6 icon-[iconamoon--shield-no-fill]"></span>`;
    case "warning":
      return `<span class="text-yellow-500 size-6 icon-[ph--warning-fill]"></span>`;
    default:
      return `<span class="text-blue-500 size-6 icon-[iconamoon--information-circle-fill]"></span>`;
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
  const toastId = `toast-${uniqueID()}`;
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
            class="inline-flex hover:text-red-500 focus:ring-2 focus:ring-white focus:ring-offset-2 focus:outline-none text-red-500/50"
            onclick="globalThis.dispatchEvent(new CustomEvent('CloseAlert', { detail: { id: '${toastId}' } }))"
          >
            <span class="sr-only">Close</span>
            <span class="icon-[iconamoon--sign-times-circle-fill] size-6"></span>
          </button>
        </div>
      </div>
    </div>
  `;
  const toastDiv = document.createElement("div");
  toastDiv.className =
    "overflow-hidden w-full max-w-sm bg-white rounded-xl ring-1 ring-black ring-opacity-5 shadow-lg pointer-events-auto text-neutral-800 border-neutral-800 shadow-neutral-800/20";
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

  toastDiv.classList.add("transform", "ease-in", "duration-200", "transition");
  toastDiv.classList.add("translate-y-0", "opacity-100");
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      toastDiv.classList.remove("translate-y-0", "opacity-100");
      toastDiv.classList.add("-translate-y-2", "opacity-0");
      setTimeout(() => {
        requestAnimationFrame(() => {
          toastDiv.remove();
        });
      }, 305);
    });
  });
}

function handleShowAlert(evt: ShowAlertEvent) {
  if (!evt.detail) {
    throw new Error("Invalid event received from server");
  }
  const { message, title, variant, duration } = evt.detail;
  if (!message || !title || !variant || !duration) {
    throw new Error("Invalid event received from server");
  }
  showAlert(message, title, variant, duration);
}

globalThis.addEventListener("ShowAlert", handleShowAlert);

function handleCloseAlert(evt: CloseAlertEvent) {
  if (!evt.detail) {
    throw new Error("Invalid event received from server");
  }
  const { id } = evt.detail;
  if (!id) {
    throw new Error("Invalid event received from server");
  }
  closeAlert(id);
}

globalThis.addEventListener("CloseAlert", handleCloseAlert);

function handleFocusInput(evt: FocusInputEvent) {
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

globalThis.addEventListener("FocusInput", handleFocusInput);

function closeModal(id: string) {
  globalThis.handleCloseModal();
  const modal = document.getElementById(id);
  if (!(modal instanceof HTMLDialogElement)) {
    return;
  }
  if (!modal) return;

  modal.classList.add("close");
  setTimeout(() => {
    modal.classList.remove("close");
    modal.close();
    setTimeout(() => modal.remove(), 1000);
  }, 155);
}

function handleConfirm(e: HTMXConfirmEvent) {
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

  globalThis.addEventListener(confirmevent, onConfirm);
  globalThis.addEventListener(cancelevent, onCancel);

  const dialog = document.createElement("confirm-dialog");
  dialog.setAttribute("dialogid", id);
  dialog.setAttribute("question", question);
  dialog.setAttribute("confirmevent", confirmevent);
  dialog.setAttribute("cancelevent", cancelevent);

  document.getElementById("main-content")?.appendChild(dialog);
  (document.getElementById(id) as HTMLDialogElement).showModal();
  globalThis.handleShowModal();
}

globalThis.addEventListener("htmx:confirm", handleConfirm);

function closeAndScroll(event: HTMXRequestEvent) {
  if (!event.detail?.target || !(event.detail.target instanceof HTMLElement)) {
    return;
  }
  if (event.detail?.target?.id === "main-content") {
    globalThis.handleCloseModal();
    window.scrollTo(0, 0);
  }
}

globalThis.addEventListener("htmx:afterSwap", closeAndScroll);

function closePopper(event: HTMXRequestEvent) {
  if (!(event.detail.target instanceof HTMLElement)) {
    return;
  }
  if (event.detail?.target?.id === "main-content") {
    document
      .querySelectorAll("[data-radix-popper-content-wrapper]")
      .forEach((el) => {
        el.remove();
      });
  }
}

globalThis.addEventListener("htmx:beforeSwap", closePopper);

function handleCloseModal(event: CloseModalEvent) {
  if (event.detail?.value) {
    globalThis.handleCloseModal();
    const modal = document.getElementById(event.detail.value);
    if (!(modal instanceof HTMLDialogElement)) {
      return;
    }

    modal.classList.add("close");
    setTimeout(() => {
      modal.classList.remove("close");
      modal.close();
    }, 155);
  }
}

globalThis.addEventListener("CloseModal", handleCloseModal);

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
        .then((res) => {
          if (res.ok) {
            window.location.href = nextLoc;
          } else {
            console.log(res);
            globalThis.dispatchEvent(
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
          globalThis.dispatchEvent(
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
      globalThis.dispatchEvent(
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
            globalThis.dispatchEvent(
              new CustomEvent("ShowAlert", {
                detail: {
                  message: "Use your passkey to login in the future!",
                  title: "Passkey Registered",
                  variant: "success",
                  duration: 3000,
                },
              }),
            );
            const passkeyCountEl = document.getElementById("passkey-count");
            if (passkeyCountEl) {
              passkeyCountEl.innerHTML = (
                parseInt(passkeyCountEl.innerHTML, 10) + 1
              ).toString();
            }
          } else {
            console.log(res);
            globalThis.dispatchEvent(
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
          globalThis.dispatchEvent(
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
      globalThis.dispatchEvent(
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

function handleShowModal() {
  document.body.style.position = "fixed;";
  document.body.style.top = `-${window.scrollY}px`;
  document.body.style.height = "100dvh";
  document.body.style.overflowY = "hidden";
}
globalThis.handleShowModal = handleShowModal;

globalThis.handleCloseModal = function () {
  const scrollY = document.body.style.top;
  document.body.style.position = "";
  document.body.style.top = "";
  document.body.style.height = "";
  document.body.style.overflowY = "";
  window.scrollTo(0, parseInt(scrollY || "0", 10) * -1);
};

globalThis.closeModal = closeModal;
