export interface HTMXRequestConfig {
  verb: string;
  path: string;
}
export interface HTMXBeforeSwapEvent extends Event {
  detail: {
    elt: HTMLElement;
    xhr: XMLHttpRequest;
    requestConfig: HTMXRequestConfig;
    shouldSwap: boolean;
    ignoreTitle: boolean;
    target: HTMLElement;
  };
}

export interface HTMXConfirmEvent extends Event {
  detail: {
    elt: HTMLElement;
    etc: object;
    issueRequest: (confirmed: boolean) => void;
    path: string;
    target: HTMLElement;
    triggeringEvent: Event;
    verb: string;
    question: string;
  };
}

export interface HTMXRequestEvent extends Event {
  detail: {
    elt: HTMLElement;
    xhr: XMLHttpRequest;
    requestConfig: HTMXRequestConfig;
    target: HTMLElement;
  };
}

export type AlertVariant = "success" | "info" | "warning" | "error";
export interface ShowAlertEvent extends Event {
  detail?: {
    message?: string;
    title?: string;
    variant?: AlertVariant;
    duration?: number;
  };
}

export interface CloseAlertEvent extends Event {
  detail?: {
    id?: string;
  };
}

export interface FocusInputEvent extends Event {
  detail?: {
    id?: string;
  };
}

export interface FinishedSpotPracticingEvent extends Event {
  detail: {
    spots: { id: string; promote: boolean; demote: boolean }[];
    durationMinutes: number;
    csrf: string;
    endpoint: string;
  };
}

export interface FinishedStartingPointPracticingEvent extends Event {
  detail: {
    durationMinutes: number;
    measuresPracticed: string;
    pieceid: string;
    csrf: string;
  };
}

export interface FinishedRepeatPracticingEvent extends Event {
  detail: {
    durationMinutes: number;
    csrf: string;
    endpoint: string;
    toStage: string;
    success: boolean;
  };
}

declare global {
  function handleCloseModal(): void;
  function handleShowModal(): void;
  interface WindowEventMap {
    "htmx:beforeSwap": HTMXBeforeSwapEvent;
    "htmx:confirm": HTMXConfirmEvent;
    "htmx:afterSwap": HTMXRequestEvent;
    "htmx:beforeSwap": HTMXRequestEvent;
    "htmx:afterRequestSwap": HTMXRequestEvent;
    ShowAlert: ShowAlertEvent;
    CloseAlert: CloseAlertEvent;
    FocusInput: FocusInputEvent;
    FinishedSpotPracticing: FinishedSpotPracticingEvent;
    FinishedStartingPointPracticing: FinishedStartingPointPracticingEvent;
    FinishedRepeatPracticing: FinishedRepeatPracticingEvent;
  }
}
export {};
