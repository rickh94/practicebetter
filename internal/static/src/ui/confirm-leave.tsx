import { useCallback, useEffect, useId, useRef } from "preact/hooks";

export function ConfirmDialog({
  onConfirm,
  onCancel,
}: {
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  const dialogid = useId();

  const close = useCallback(() => {
    globalThis.handleCloseModal();
    if (ref.current) {
      ref.current?.classList.add("close");
      setTimeout(() => {
        ref.current?.classList.remove("close");
        ref.current?.close();
        ref.current?.remove();
      }, 75);
    }
  }, [ref]);

  const handleConfirm = useCallback(() => {
    onConfirm();
    close();
  }, [onConfirm, close]);
  const handleCancel = useCallback(() => {
    onCancel();
    close();
  }, [onCancel, close]);
  useEffect(() => {
    if (ref.current) {
      globalThis.handleShowModal();
      ref.current.showModal();
    }
    return function () {
      close();
    };
  }, [close]);
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
        <button
          onClick={handleCancel}
          className="action-button sky focusable flex-grow"
        >
          Stay
        </button>
        <button
          onClick={handleConfirm}
          className="action-button amber focusable flex-grow"
        >
          Leave
        </button>
      </section>
    </dialog>
  );
}
