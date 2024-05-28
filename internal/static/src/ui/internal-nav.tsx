import { type NavItem, cn } from "../common";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import * as htmx from "htmx.org/dist/htmx";
import { useCallback, useId } from "preact/hooks";

export function InternalNav({ activepath }: { activepath: string }) {
  const id = useId();
  const links: NavItem[] = [
    {
      href: "/library",
      label: "Library",
      icon: <span className="icon-[custom--music-note-screen] ml-1 size-6" />,
    } as const,
    {
      href: "/library/plans",
      label: "Practice Plans",
      icon: <span className="icon-[iconamoon--calendar-2-thin] ml-1 size-6" />,
    } as const,
    {
      href: "/library/pieces",
      label: "Pieces",
      icon: <span className="icon-[custom--music-folder] ml-1 size-6" />,
    } as const,
    {
      href: "/library/scales",
      label: "Scales",
      icon: <span className="icon-[ph--steps-thin] ml-1 size-6" />,
    } as const,
    {
      href: "/library/reading",
      label: "Sight Reading",
      icon: (
        <span className="icon-[fluent--reading-mode-mobile-20-regular] ml-1 size-6" />
      ),
    } as const,
    {
      href: "/practice",
      label: "Practice Tools",
      icon: <span className="icon-[iconamoon--playlist-thin] ml-1 size-6" />,
    } as const,
    {
      href: "/auth/me",
      label: "Account",
      icon: (
        <span
          className="icon-[iconamoon--profile-circle-thin] ml-1 size-6"
          aria-hidden="true"
        />
      ),
    } as const,
  ];

  const processLinks = useCallback(
    (open: boolean) => {
      if (!open) {
        return;
      }
      document
        .querySelectorAll(`#${id} a[data-radix-collection-item]`)
        .forEach((el) => {
          if (el) {
            htmx.process(el);
          }
        });
    },
    [id],
  );

  return (
    <DropdownMenu.Root onOpenChange={processLinks}>
      <DropdownMenu.Trigger asChild>
        <button
          className="focusable flex h-14 items-center justify-center gap-x-1.5 rounded-xl border border-neutral-300 bg-neutral-50 px-6 text-neutral-700 shadow-sm shadow-neutral-900/20 hover:border-neutral-500 hover:bg-neutral-100 hover:shadow hover:shadow-neutral-900/30"
          aria-label="Menu"
        >
          <div className="sr-only">Open Nav Menu</div>
          <span
            className="icon-[iconamoon--menu-burger-horizontal-thin] -ml-2 size-6"
            aria-hidden="true"
          />
          <span className="font-medium">Menu</span>
        </button>
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          as="nav"
          side="bottom"
          align="start"
          sideOffset={5}
          className="dropdown w-64 origin-top-left rounded-lg bg-white shadow-lg duration-200 focus-within:outline-none focus:outline-none"
          id={id}
        >
          {links.map(({ href, label, icon }) => (
            <DropdownMenu.Item asChild key={href}>
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
                    ? "bg-neutral-700/10 font-semibold text-neutral-800"
                    : "text-neutral-800 hover:bg-neutral-800/10 focus-visible:bg-neutral-800/10",
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
