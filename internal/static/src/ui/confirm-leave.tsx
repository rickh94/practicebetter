import { useCallback, useEffect, useId, useRef } from "preact/hooks";
import { SkyButton, WarningButton } from "./buttons";

export function ConfirmDialog({
  onConfirm,
  onCancel,
}: {
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  const dialogid = useId();

  const close = useCallback(
    function () {
      if (ref.current) {
        if (!ref.current) return;
        ref.current.classList.add("close");

        requestAnimationFrame(function () {
          requestAnimationFrame(function () {
            ref.current.classList.remove("close");
            // @ts-ignore
            modal.close();
            setTimeout(() => ref.current.remove(), 1000);
          });
        });
      }
    },
    [ref],
  );

  const handleConfirm = useCallback(
    function () {
      onConfirm();
      close();
    },
    [onConfirm, close],
  );
  const handleCancel = useCallback(
    function () {
      onCancel();
      close();
    },
    [onCancel, close],
  );
  useEffect(function () {
    if (ref.current) {
      ref.current.showModal();
    }
    return function () {
      close();
    };
  }, []);
  return (
    <dialog
      id={dialogid}
      ref={ref}
      aria-labelledby={`${dialogid}-title`}
      className="flex flex-col gap-2 bg-gradient-to-t from-neutral-50 to-[#fff9ee] px-4 py-4 text-left sm:max-w-xl"
    >
      <header className="mt-2 text-center sm:text-left">
        <h3
          id={`${dialogid}-title`}
          class="text-2xl font-semibold leading-6 text-neutral-900"
        >
          Leaving Page
        </h3>
      </header>
      <main className="mt-2 text-left text-neutral-800">
        <p>
          You have unsaved changes on this page. Are you sure you want to leave?
        </p>
      </main>
      <section className="mt-4 flex w-full gap-2">
        <WarningButton onClick={handleCancel} grow>
          Stay
        </WarningButton>
        <SkyButton onClick={handleConfirm} grow>
          Leave
        </SkyButton>
      </section>
    </dialog>
  );
}
