import { type JSX } from "preact/jsx-runtime";
import { cn } from "../common";
import { RandomBoxesIcon, RepeatIcon, ShuffleIcon } from "./icons";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import * as htmx from "htmx.org/dist/htmx";

const links = new Map<string, { label: string; icon: JSX.Element }>([
  [
    "/practice/repeat",

    {
      label: "Repeat Practice",
      icon: <RepeatIcon className="ml-1 size-5" />,
    },
  ],
  [
    "/practice/random-single",
    { label: "Random Spots", icon: <ShuffleIcon className="ml-1 size-5" /> },
  ],
  // ["/practice/random-sequence", "Randomized Sequence"],
  [
    "/practice/starting-point",
    {
      label: "Random Starting Point",
      icon: <RandomBoxesIcon className="ml-1 size-5" />,
    },
  ],
]);

function processLinks() {
  document.querySelectorAll("a[data-radix-collection-item]").forEach((el) => {
    console.log("processing nav menu link");
    htmx.process(el);
  });
}

export function PracticeToolNav({ activepath }: { activepath: string }) {
  return (
    <DropdownMenu.Root onOpenChange={processLinks}>
      <DropdownMenu.Trigger asChild>
        <button className="focusable mx-auto flex h-14 items-center justify-center gap-x-1.5 rounded-xl border border-neutral-300 bg-neutral-50 px-6 text-neutral-700 drop-shadow-sm transition duration-200 ease-in hover:border-neutral-500 hover:bg-neutral-200 hover:drop-shadow-md">
          <div className="sr-only">Open Practice Tools Menu</div>
          <span className="icon-[iconamoon--playlist-thin] -ml-2 size-8 text-neutral-800" />
          {links.get(activepath) ? (
            <h1 className="text-xl font-semibold tracking-tight text-neutral-800 sm:text-2xl">
              {links.get(activepath)?.label ?? "Something went wrong"}
            </h1>
          ) : (
            <span className="font-semibold text-neutral-700">
              Practice Tools
            </span>
          )}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          side="bottom"
          align="start"
          sideOffset={5}
          className="dropdown w-64 origin-top-left rounded-lg bg-white shadow-lg focus-within:outline-none focus:outline-none"
        >
          {[...links.entries()].map(([href, info]) => (
            <DropdownMenu.Item asChild key={href}>
              <a
                href={href}
                onClick={(e) => e.preventDefault()}
                hx-get={href}
                hx-swap="outerHTML transition:true"
                hx-push-url="true"
                hx-target="#main-content"
                className={cn(
                  "flex w-full items-center gap-1 px-2 py-3 text-lg text-neutral-800 first:rounded-t-lg last:rounded-b-lg focus:outline-none",
                  activepath === href
                    ? "bg-neutral-700/10 font-bold"
                    : "font-medium hover:bg-neutral-800/10 focus-visible:bg-neutral-800/10",
                )}
              >
                {info.icon} {info.label}
              </a>
            </DropdownMenu.Item>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
