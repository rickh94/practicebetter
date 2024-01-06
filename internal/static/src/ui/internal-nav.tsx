import { NavItem, cn } from "../common";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { NoteSheetIcon, PlayListIcon } from "./icons";

const navItemIconClasses = "ml-1 size-5" as const;

export function InternalNav({
  activeplanid,
  activepath,
}: {
  activeplanid: string;
  activepath: string;
}) {
  const links: NavItem[] = [
    {
      href: "/library",
      label: "Library",
      icon: (
        <span
          className={cn(
            "icon-[solar--music-note-slider-bold-duotone]",
            navItemIconClasses,
          )}
        />
      ),
    } as const,
    {
      href: "/library/plans",
      label: "Practice Plans",
      icon: (
        <span
          className={cn(
            "icon-[solar--clipboard-list-bold]",
            navItemIconClasses,
          )}
        />
      ),
    } as const,
    {
      href: "/library/pieces",
      label: "Pieces",
      icon: <NoteSheetIcon className={navItemIconClasses} />,
    } as const,
    {
      href: "/practice",
      label: "Practice Tools",
      icon: <PlayListIcon className={navItemIconClasses} />,
    } as const,
    {
      href: "/auth/me",
      label: "Account",
      icon: (
        <span
          className={cn(
            "icon-[heroicons--user-circle-solid]",
            navItemIconClasses,
          )}
          aria-hidden="true"
        />
      ),
    } as const,
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
        <button
          className="focusable inline-flex h-14 w-full items-center justify-center gap-x-1.5 rounded-xl bg-neutral-700/10 px-6 py-4 shadow-sm transition duration-200 hover:bg-neutral-700/20"
          aria-label="Customise options"
        >
          <div className="sr-only">Open Nav Menu</div>
          <span
            className="icon-[heroicons--bars-3-center-left-solid] -ml-2 size-6 text-neutral-800"
            aria-hidden="true"
          />
          <span className="font-medium text-neutral-800">Menu</span>
        </button>
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          as="nav"
          side="bottom"
          align="start"
          sideOffset={5}
          className="w-64 origin-top-left rounded-lg bg-white shadow-lg duration-200 animate-in fade-in zoom-in-95 focus-within:outline-none focus:outline-none"
        >
          {links.map(({ href, label, icon }) => (
            <DropdownMenu.Item asChild>
              <a
                href={href}
                onClick={(e) => e.preventDefault()}
                hx-get={href}
                hx-swap="outerHTML transition:true"
                hx-push-url="true"
                hx-target="#main-content"
                className={cn(
                  "flex w-full items-center gap-2 px-2 py-3 text-lg first:rounded-t-lg last:rounded-b-lg focus:outline-none",
                  href === activepath
                    ? "bg-neutral-700/10 font-bold text-neutral-800"
                    : "font-medium text-neutral-800 hover:bg-neutral-800/10 focus-visible:bg-neutral-800/10",
                )}
              >
                {icon}
                {label}
              </a>
            </DropdownMenu.Item>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
