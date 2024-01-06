import {
  PlayListIcon,
  RandomBoxesIcon,
  RepeatIcon,
  ShuffleIcon,
} from "./icons";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";

export function PracticeMenu({ pieceid }: { pieceid: string }) {
  const links = [
    {
      href: `/library/pieces/${pieceid}/practice/repeat`,
      label: "Repeat Practice",
      icon: <RepeatIcon className="ml-1 size-5" />,
    },
    {
      href: `/library/pieces/${pieceid}/practice/random-single`,
      label: "Random Spots",
      icon: <ShuffleIcon className="ml-1 size-5" />,
    },
    {
      href: `/library/pieces/${pieceid}/practice/starting-point`,
      label: "Random Starting Point",
      icon: <RandomBoxesIcon className="ml-1 size-5" />,
    },
  ];

  function processLinks() {
    document.querySelectorAll("a[data-radix-collection-item]").forEach((el) => {
      // @ts-ignore
      if (htmx) {
        console.log("processing");
        // @ts-ignore
        htmx.process(el);
      }
    });
  }

  return (
    <DropdownMenu.Root onOpenChange={processLinks}>
      <DropdownMenu.Trigger asChild>
        <button className="focusable action-button bg-violet-700/10 text-violet-800 transition-all duration-200 ease-out hover:bg-violet-700/20">
          <PlayListIcon className="-ml-1 size-5" />
          Practice
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          side="bottom"
          align="end"
          sideOffset={5}
          className="w-64 origin-top-right rounded-lg bg-white shadow-lg duration-200 animate-in fade-in zoom-in-95 focus-within:outline-none focus:outline-none"
        >
          {links.map((link) => (
            <DropdownMenu.Item asChild>
              <a
                href={link.href}
                onClick={(e) => e.preventDefault()}
                hx-get={link.href}
                hx-swap="outerHTML transition:true"
                hx-push-url="true"
                hx-target="#main-content"
                className="flex w-full items-center gap-1 px-2 py-3 text-lg font-medium text-violet-950 first:rounded-t-lg last:rounded-b-lg hover:bg-violet-800/10 focus:outline-none focus-visible:bg-violet-800/10"
              >
                {link.icon} {link.label}
              </a>
            </DropdownMenu.Item>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
