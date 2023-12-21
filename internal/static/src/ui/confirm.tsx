import { useCallback } from "preact/hooks";
import { SkyButton, WarningButton } from "./buttons";

export function ConfirmDialog({
  question,
  confirmevent,
  cancelevent,
  dialogid,
}: {
  dialogid: string;
  question: string;
  confirmevent: string;
  cancelevent: string;
}) {
  const onConfirm = useCallback(
    function () {
      document.dispatchEvent(new CustomEvent(confirmevent));
    },
    [confirmevent],
  );
  const onCancel = useCallback(
    function () {
      document.dispatchEvent(new CustomEvent(cancelevent));
    },
    [cancelevent],
  );
  return (
    <dialog
      id={dialogid}
      aria-labelledby={`${dialogid}-title`}
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
    >
      <header className="mt-2 text-center sm:text-left">
        <h3
          id={`${dialogid}-title`}
          class="text-2xl font-semibold leading-6 text-neutral-900"
        >
          Please Confirm
        </h3>
      </header>
      <main className="mt-2 text-left text-neutral-800">
        <p>{question}</p>
      </main>
      <section className="mt-4 flex w-full flex-wrap gap-2">
        <WarningButton onClick={onCancel} grow>
          Cancel
        </WarningButton>
        <SkyButton onClick={onConfirm} grow>
          Confirm
        </SkyButton>
      </section>
    </dialog>
  );
}
