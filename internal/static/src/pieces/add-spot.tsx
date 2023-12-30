import { yupResolver } from "@hookform/resolvers/yup";
import { useForm } from "react-hook-form";
import { SpotFormData, spotFormData } from "../validators";
import SpotFormFields from "./spot-form";
import { useCallback, useEffect, useState } from "preact/hooks";

export function AddSpotForm({
  pieceid,
  csrf,
  initialspotcount,
}: {
  pieceid: string;
  csrf: string;
  initialspotcount: string;
}) {
  const [isUpdating, setIsUpdating] = useState(false);
  const [nextSpotIdx, setNextSpotIdx] = useState(
    parseInt(initialspotcount) + 1 || 1,
  );
  const {
    handleSubmit,
    formState,
    formState: { isDirty, dirtyFields },
    setValue,
    reset,
    register,
    watch,
  } = useForm<SpotFormData>({
    mode: "onChange",
    reValidateMode: "onBlur",
    resolver: yupResolver(spotFormData),
    defaultValues: {
      name: "",
      idx: nextSpotIdx,
      measures: "",
      audioPromptUrl: "",
      textPrompt: "",
      notesPrompt: "",
      imagePromptUrl: "",
      stage: "repeat",
    },
  });

  async function onSubmit(data: SpotFormData, e: Event) {
    e.preventDefault();
    setIsUpdating(true);
    // @ts-ignore
    await htmx.ajax("POST", `/library/pieces/${pieceid}/spots`, {
      values: data,
      target: "#spot-list",
      swap: "beforeend transition:true",
      headers: {
        "X-CSRF-Token": csrf,
      },
    });
    reset(
      {
        name: "",
        idx: nextSpotIdx + 1,
        measures: "",
        audioPromptUrl: "",
        textPrompt: "",
        notesPrompt: "",
        imagePromptUrl: "",
        stage: "repeat",
      },
      { keepDirty: false },
    );
    setIsUpdating(false);
    document.getElementById("spot-count").textContent = `(${nextSpotIdx})`;
    setNextSpotIdx(nextSpotIdx + 1);
  }

  const onBeforeUnload = useCallback(function (e: BeforeUnloadEvent) {
    e.preventDefault();
    e.returnValue = true;
  }, []);

  const onBeforeHtmxRequest = useCallback(function (e: CustomEvent) {
    if (
      e.detail.requestConfig.verb === "post" &&
      e.detail.requestConfig.path === `/library/pieces/${pieceid}/spots`
    ) {
      return;
    }
    if (confirm("You have unsaved changes. Are you sure you want to leave?")) {
      return;
    }
    e.preventDefault();
  }, []);

  useEffect(
    function () {
      if (isDirty && Object.keys(dirtyFields).length > 0) {
        window.addEventListener("beforeunload", onBeforeUnload);
        document.addEventListener("htmx:beforeRequest", onBeforeHtmxRequest);
        return function () {
          window.removeEventListener("beforeunload", onBeforeUnload);
          document.removeEventListener(
            "htmx:beforeRequest",
            onBeforeHtmxRequest,
          );
        };
      } else {
        window.removeEventListener("beforeunload", onBeforeUnload);
        document.removeEventListener("htmx:beforeRequest", onBeforeHtmxRequest);
      }
    },
    [isDirty, Object.keys(dirtyFields)],
  );

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      noValidate
      hx-boost="false"
      id="add-spot-form"
    >
      <SpotFormFields
        csrf={csrf}
        isUpdating={isUpdating}
        formState={formState}
        setValue={setValue}
        backTo={`/library/pieces/${pieceid}`}
        register={register}
        watch={watch}
      />
    </form>
  );
}
