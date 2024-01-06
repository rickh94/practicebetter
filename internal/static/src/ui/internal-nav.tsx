import { NavItem, cn } from "../common";
import {
  Bars3CenterLeftIcon,
  ClipboardDocumentListIcon,
  RectangleStackIcon,
  UserCircleIcon,
} from "@heroicons/react/24/solid";
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
      icon: <RectangleStackIcon className={navItemIconClasses} />,
    } as const,
    {
      href: "/library/plans",
      label: "Practice Plans",
      icon: <ClipboardDocumentListIcon className={navItemIconClasses} />,
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
      icon: <UserCircleIcon className={navItemIconClasses} />,
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
          <Bars3CenterLeftIcon
            className="-ml-2 size-6 text-neutral-800"
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

/*
 *
      <Menu as="nav" className="relative inline-block text-left">
        <Menu.Button
          className={cn(
            "focusable inline-flex h-14 w-full items-center justify-center gap-x-1.5 rounded-xl px-6 py-4 shadow-sm transition duration-200",
            activeplanid
              ? "bg-violet-700/30 hover:bg-violet-700/40"
              : "bg-neutral-700/10 hover:bg-neutral-700/20",
          )}
        >
          {({ open }) => (
            <>
              <CrossFadeContentFast
                id={open ? "open" : "closed"}
                component={
                  open ? (
                    <>
                      <div className="sr-only">Close Nav Menu</div>
                      <XMarkIcon
                        className="-ml-2 size-6 text-neutral-800"
                        aria-hidden="true"
                      />
                    </>
                  ) : (
                    <>
                      <div className="sr-only">Open Nav Menu</div>
                      <Bars3CenterLeftIcon
                        className="-ml-2 size-6 text-neutral-800"
                        aria-hidden="true"
                      />
                    </>
                  )
                }
              />
              <span className="font-medium text-neutral-800">Menu</span>
            </>
          )}
        </Menu.Button>
        <Transition
          as={Fragment}
          enter="transition ease-out duration-100"
          enterFrom="transform opacity-0 scale-95"
          enterTo="transform opacity-100 scale-100"
          leave="transition ease-in duration-75"
          leaveFrom="transform opacity-100 scale-100"
          leaveTo="transform opacity-0 scale-95"
        >
          <Menu.Items
            // @ts-ignore
            className="absolute left-0 z-50 mt-2 w-64 origin-top-left rounded-lg bg-white shadow-lg focus-within:outline-none focus:outline-none"
            as="nav"
          >
            <ul className="flex flex-col gap-0">
              {links.map(({ href, label, icon, highlight }) => (
                <Menu.Item
                  key={href}
                  as="li"
                  // @ts-ignore
                  className={cn(
                    "focusable w-full text-lg first:rounded-t-lg last:rounded-b-lg",
                    highlight
                      ? {
                          "bg-violet-700/30 font-bold text-violet-800":
                            href === activepath,
                          "font-medium text-violet-800 hover:bg-violet-800/30":
                            href !== activepath,
                        }
                      : {
                          "bg-neutral-700/10 font-bold text-neutral-800":
                            href === activepath,
                          "font-medium text-neutral-800 hover:bg-neutral-800/10":
                            href !== activepath,
                        },
                  )}
                >
                  <Link
                    href={href}
                    target="#main-content"
                    className="focusable flex h-full w-full items-center gap-2 px-2 py-3"
                  >
                    {icon}
                    {label}
                  </Link>
                </Menu.Item>
              ))}
            </ul>
          </Menu.Items>
        </Transition>
      </Menu>
*/
