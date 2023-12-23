import { Menu, Transition } from "@headlessui/react";
import { PlayIcon } from "@heroicons/react/20/solid";
import { Fragment } from "react";
import { Link } from "./links";
import {
  NoteListIcon,
  PlayListIcon,
  RandomBoxesIcon,
  RepeatIcon,
  ShuffleIcon,
} from "./icons";
import { XMarkIcon } from "@heroicons/react/24/solid";

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

  return (
    <>
      {/*
    // @ts-ignore */}
      <Menu as="div" className="relative inline-block text-left">
        {/*
      // @ts-ignore */}
        <Menu.Button className="focusable action-button bg-green-700/10 text-green-800 transition-all duration-200 ease-out hover:bg-green-700/20">
          {({ open }) => (
            <>
              {open ? (
                <XMarkIcon className="-ml-1 size-5" />
              ) : (
                <PlayListIcon className="-ml-1 size-5" />
              )}
              Practice
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
            className="absolute right-0 z-10 mt-2 w-64 origin-top-right rounded-lg bg-white shadow-lg ring-1 ring-black  ring-opacity-5 backdrop-blur focus:outline-none sm:left-0 sm:origin-top-left"
            as="nav"
          >
            <ul className="flex flex-col gap-0">
              {links.map((link) => (
                <Menu.Item
                  key={link.href}
                  as="li"
                  // @ts-ignore
                  className="w-full text-lg font-medium text-green-950 first:rounded-t-lg last:rounded-b-lg hover:bg-green-500/20"
                >
                  <Link
                    href={link.href}
                    className="flex h-full w-full items-center gap-1 px-2 py-3"
                  >
                    {link.icon}
                    {link.label}
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
