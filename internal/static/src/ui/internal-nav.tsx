import { NavItem, cn } from "../common";
import { Bars3CenterLeftIcon, XMarkIcon } from "@heroicons/react/20/solid";
import { CrossFadeContentFast } from "../ui/transitions";
import { Menu, Transition } from "@headlessui/react";
import { Fragment } from "preact/jsx-runtime";
import { Link } from "./links";
import { useMemo } from "preact/hooks";
import {
  BookOpenIcon,
  DocumentCheckIcon,
  DocumentMagnifyingGlassIcon,
  DocumentTextIcon,
  HomeIcon,
  QueueListIcon,
} from "@heroicons/react/24/solid";

const navItemIconClasses = "ml-1 size-5" as const;

export function InternalNav({
  activeplanid,
  activepath,
}: {
  activeplanid: string;
  activepath: string;
}) {
  const links: NavItem[] = useMemo(() => {
    const linkList: NavItem[] = [
      {
        href: "/library",
        label: "Library",
        icon: <BookOpenIcon className={navItemIconClasses} />,
      } as const,
      {
        href: "/library/plans",
        label: "Practice Plans",
        icon: <DocumentTextIcon className={navItemIconClasses} />,
      } as const,
      {
        href: "/library/pieces",
        label: "Pieces",
        icon: <DocumentMagnifyingGlassIcon className={navItemIconClasses} />,
      } as const,
      {
        href: "/",
        label: "Home",
        icon: <HomeIcon className={navItemIconClasses} />,
      } as const,
    ];
    if (activeplanid) {
      linkList.unshift({
        href: `/library/plans/${activeplanid}`,
        label: "Current Practice Plan",
        icon: <DocumentCheckIcon className={navItemIconClasses} />,
        highlight: true,
      });
    }
    return linkList;
  }, [activeplanid]);

  return (
    <>
      {/*
      // @ts-ignore */}
      <Menu as="nav" className="relative inline-block text-left">
        {/*
        // @ts-ignore */}
        <Menu.Button
          className={cn(
            "focusable inline-flex h-14 w-full items-center justify-center gap-x-1.5 rounded-xl px-6 py-4 shadow-sm",
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
                        className={cn(
                          "-ml-2 size-6",
                          activeplanid ? "text-violet-900" : "text-neutral-800",
                        )}
                        aria-hidden="true"
                      />
                    </>
                  ) : (
                    <>
                      <div className="sr-only">Open Nav Menu</div>
                      <Bars3CenterLeftIcon
                        className={cn(
                          "-ml-2 size-6",
                          activeplanid ? "text-violet-900" : "text-neutral-800",
                        )}
                        aria-hidden="true"
                      />
                    </>
                  )
                }
              />
              <span
                className={cn(
                  "font-medium",
                  activeplanid ? "text-violet-900" : "text-neutral-800",
                )}
              >
                Menu
              </span>
              <span
                className={cn(
                  "font-medium",
                  activeplanid ? "hidden text-violet-900 sm:inline" : "hidden",
                )}
              >
                - Practicing
              </span>
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
            className="absolute left-0 z-50 mt-2 w-64 origin-top-left rounded-lg bg-white shadow-lg  ring-1 ring-black ring-opacity-5 backdrop-blur focus:outline-none"
            as="nav"
          >
            <ul className="flex flex-col gap-0">
              {links.map(({ href, label, icon, highlight }) => (
                <Menu.Item
                  key={href}
                  as="li"
                  // @ts-ignore
                  className={cn(
                    "w-full text-lg first:rounded-t-lg last:rounded-b-lg",
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
                    className="flex h-full w-full items-center gap-2 px-2 py-3"
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
    </>
  );
}
