function maybeShowDialog() {
  // check for url query value recommended
  const recommended = new URLSearchParams(window.location.search).get(
    "recommend",
  );
  if (recommended === "1") {
    const dialog = document.getElementById("recommend-dialog");
    if (dialog instanceof HTMLDialogElement) {
      dialog.showModal();
      globalThis.handleShowModal();
    }
  }
}

document.addEventListener("DOMContentLoaded", () => {
  maybeShowDialog();
});

document.addEventListener("htmx:afterSettle", () => {
  maybeShowDialog();
});
