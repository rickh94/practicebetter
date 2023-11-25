import { CheckCircleIcon } from "@heroicons/react/20/solid";

export function RadioBox({
  text,
  setSelected,
  selected,
  value,
  name,
}: {
  text: string;
  setSelected: () => void;
  selected: boolean;
  value: string;
  name: string;
}) {
  return (
    <label
      htmlFor={`${name}-${value}`}
      className={`focusable relative flex cursor-pointer rounded-xl py-3 pl-4 pr-2 ${
        selected ? " bg-neutral-700/20 shadow" : " bg-neutral-700/10"
      }`}
    >
      <input
        type="radio"
        name={name}
        id={`${name}-${value}`}
        value={value}
        className="sr-only"
        checked={selected}
        onChange={(e) => e.currentTarget.checked && setSelected()}
        aria-labelledby={`${name}-${value}-label`}
      />
      <span className="flex flex-1">
        <span className="flex-col">
          <span
            className={`block text-sm text-neutral-800 ${
              selected ? "font-bold" : "font-semibold"
            }`}
            id={`${name}-${value}-label`}
          >
            {text}
          </span>
        </span>
      </span>
      <CheckCircleIcon
        className={`ml-1 h-5 w-5 text-neutral-800 ${
          selected ? "" : " invisible"
        }`}
        aria-hidden="true"
      />
      <span
        className={`pointer-events-none absolute -inset-px rounded-xl border-2${
          selected ? " border-neutral-800" : " border-transparent"
        }`}
        aria-hidden="true"
      />
    </label>
  );
}
