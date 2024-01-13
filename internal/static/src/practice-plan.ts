type UpdatePlanProgressEvent = Event & {
  detail?: {
    completed: number;
    total: number;
  };
};

function handleUpdatePlanProgressEvent(evt: UpdatePlanProgressEvent) {
  console.log("update progress");
  if (!evt.detail) {
    throw new Error("Invalid event received from server");
  }
  const { completed, total } = evt.detail;
  if (isNaN(completed) || isNaN(total)) {
    throw new Error("Invalid event received from server");
  }
  const pb = document.getElementById("plan-progress-bar");
  if (!(pb instanceof HTMLProgressElement)) {
    throw new Error("Incorrect element");
  }
  pb.value = completed;
  pb.max = total;
  pb.innerText = `${completed}/${total}`;
}
document.addEventListener("UpdatePlanProgress", handleUpdatePlanProgressEvent);
