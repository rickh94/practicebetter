import { cn } from "../common";
import { Bars3CenterLeftIcon, XMarkIcon } from "@heroicons/react/20/solid";
import { CrossFadeContentFast } from "../ui/transitions";
import { Menu, Transition } from "@headlessui/react";
import { Fragment } from "preact/jsx-runtime";
import { Link } from "./links";
import { RandomBoxesIcon, RepeatIcon, ShuffleIcon } from "./icons";

const links = new Map<string, { label: string; icon: preact.JSX.Element }>([
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

export function PracticeToolNav({ activepath }: { activepath: string }) {
  return (
    <>
      {/*
      // @ts-ignore */}
      <Menu as="div" className="relative inline-block text-left">
        {/*
        // @ts-ignore */}
        <Menu.Button className="focusable inline-flex h-14 w-full items-center justify-center gap-x-1.5 rounded-xl bg-neutral-700/10 px-6 py-4 shadow-sm hover:bg-neutral-700/20">
          {({ open }) => (
            <>
              <CrossFadeContentFast
                id={open ? "open" : "closed"}
                component={
                  open ? (
                    <>
                      <div className="sr-only">Close Practice Tools Menu</div>
                      <XMarkIcon
                        className="-ml-2 size-6 text-neutral-800"
                        aria-hidden="true"
                      />
                    </>
                  ) : (
                    <>
                      <div className="sr-only">Open Practice Tools Menu</div>
                      <Bars3CenterLeftIcon
                        className="-ml-2 size-6 text-neutral-800"
                        aria-hidden="true"
                      />
                    </>
                  )
                }
              />
              {links.get(activepath) ? (
                <h1 className="text-xl font-semibold tracking-tight text-neutral-800 sm:text-2xl">
                  {links.get(activepath).label}
                </h1>
              ) : (
                <span className="font-semibold text-neutral-700">
                  Practice Tools
                </span>
              )}
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
            className="absolute left-0 z-50 mt-2 w-64 origin-top-left rounded-lg bg-[#fffaf0]/90 shadow-lg  ring-1 ring-black ring-opacity-5 backdrop-blur focus:outline-none"
            as="nav"
          >
            <ul className="flex flex-col gap-0">
              {[...links.entries()].map(([href, info]) => (
                <Menu.Item
                  key={href}
                  as="li"
                  // @ts-ignore
                  className={cn(
                    "w-full text-lg text-neutral-800 first:rounded-t-lg last:rounded-b-lg",
                    {
                      "bg-neutral-700/10 font-bold": href === activepath,
                      "font-medium hover:bg-neutral-800/10":
                        href !== activepath,
                    },
                  )}
                >
                  <Link
                    href={href}
                    target="#main-content"
                    className="flex h-full w-full items-center gap-1 px-2 py-3"
                  >
                    {info.icon}
                    {info.label}
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
