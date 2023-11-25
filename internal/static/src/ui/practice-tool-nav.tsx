import { cn } from "../common";
import { Bars3CenterLeftIcon, XMarkIcon } from "@heroicons/react/20/solid";
import { CrossFadeContentFast } from "../ui/transitions";
import { Menu, Transition } from "@headlessui/react";
import { Fragment } from "preact/jsx-runtime";
import { Link } from "./links";

const links = new Map([
  ["/practice/random-single", "Random Spots"],
  ["/practice/random-sequence", "Randomized Sequence"],
  ["/practice/repeat", "Repeat Practice"],
  ["/practice/starting-point", "Random Starting Point"],
]);

export function PracticeToolNav({ activepath }: { activepath: string }) {
  return (
    <>
      {/*
      // @ts-ignore */}
      <Menu as="div" className="relative -mt-4 inline-block text-left">
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
                        className="-ml-2 h-6 w-6 text-neutral-800"
                        aria-hidden="true"
                      />
                    </>
                  ) : (
                    <>
                      <div className="sr-only">Open Practice Tools Menu</div>
                      <Bars3CenterLeftIcon
                        className="-ml-2 h-6 w-6 text-neutral-800"
                        aria-hidden="true"
                      />
                    </>
                  )
                }
              />
              {links.get(activepath) ? (
                <h1 className="text-xl font-semibold tracking-tight text-neutral-800 sm:text-2xl">
                  {links.get(activepath)}
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
            className="absolute left-0 z-50 mt-2 w-56 origin-top-left rounded-lg bg-[#fffaf0]/90 shadow-lg  ring-1 ring-black ring-opacity-5 backdrop-blur focus:outline-none"
            as="nav"
          >
            <ul className="flex flex-col gap-0">
              {[...links.entries()].map(([href, label]) => (
                <Menu.Item
                  key={href}
                  as="li"
                  // @ts-ignore
                  className={cn(
                    "w-full text-lg text-neutral-800 first:rounded-t-lg last:rounded-b-lg",
                    {
                      "bg-neutral-700/10 font-bold": href === activepath,
                      "font-semibold hover:bg-neutral-800/10":
                        href !== activepath,
                    },
                  )}
                >
                  <Link
                    href={href}
                    target="#main-content"
                    className="block h-full w-full px-2 py-2"
                  >
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

/*
 *
    <Menu as="div" className="relative -mt-4 inline-block text-left">
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
                      className="-ml-2 h-6 w-6 text-neutral-800"
                      aria-hidden="true"
                    />
                  </>
                ) : (
                  <>
                    <div className="sr-only">Open Practice Tools Menu</div>
                    <Bars3CenterLeftIcon
                      className="-ml-2 h-6 w-6 text-neutral-800"
                      aria-hidden="true"
                    />
                  </>
                )
              }
            />
            {!!current?.label ? (
              <h1 className="text-xl font-semibold tracking-tight text-neutral-800 sm:text-2xl">
                {current.label}
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
          className="absolute left-0 z-10 mt-2 w-56 origin-top-left rounded-lg bg-[#fffaf0]/90 shadow-lg  ring-1 ring-black ring-opacity-5 backdrop-blur focus:outline-none"
          as="nav"
        >
          <ul className="flex flex-col gap-0">
            {links.map((link) => (
              <Menu.Item
                key={link.href}
                as="li"
                className={cn(
                  "w-full text-lg text-neutral-800 first:rounded-t-lg last:rounded-b-lg",
                  {
                    "bg-neutral-700/10 font-bold": link.href === pathname,
                    "font-semibold hover:bg-neutral-800/10":
                      link.href !== pathname,
                  },
                )}
              >
                <a
                  href={link.href}
                  className="block h-full w-full px-2 py-2"
                  hx-get=""
                >
                  {link.label}
                </a>
              </Menu.Item>
            ))}
          </ul>
        </Menu.Items>
      </Transition>
    </Menu>

    <Menu as="div" className="relative -mt-4 inline-block text-left">
      <Menu.Button className="focusable inline-flex h-14 w-full items-center justify-center gap-x-1.5 rounded-xl bg-neutral-700/10 px-6 py-4 shadow-sm hover:bg-neutral-700/20">
      </Menu.Button>
      </Menu>
    <div
      className="relative -mt-4 inline-block text-left"
      id="practice-tool-nav"
    >
      <div>
        <button
          type="button"
          className="focusable inline-flex h-14 w-full items-center justify-center gap-x-1.5 rounded-xl bg-neutral-700/10 px-6 py-4 shadow-sm hover:bg-neutral-700/20"
          id="menu-button"
          aria-expanded="true"
          aria-haspopup="true"
          onClick={() => setOpen(!open)}
        >
          <CrossFadeContentFast
            id={open ? "open" : "closed"}
            component={
              open ? (
                <>
                  <div className="sr-only">Close Practice Tools Menu</div>
                  <XMarkIcon
                    className="-ml-2 h-6 w-6 text-neutral-800"
                    aria-hidden="true"
                  />
                </>
              ) : (
                <>
                  <div className="sr-only">Open Practice Tools Menu</div>
                  <Bars3CenterLeftIcon
                    className="-ml-2 h-6 w-6 text-neutral-800"
                    aria-hidden="true"
                  />
                </>
              )
            }
          />
          {!!current?.label ? (
            <h1 className="text-xl font-semibold tracking-tight text-neutral-800 sm:text-2xl">
              {current.label}
            </h1>
          ) : (
            <span className="font-semibold text-neutral-700">
              Practice Tools
            </span>
          )}
        </button>
      </div>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
        show={open}
      >
        <nav
          className="absolute right-0 z-10 mt-2 w-56 origin-top-left rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
          role="menu"
          aria-orientation="vertical"
          aria-labelledby="menu-button"
          tabindex={-1}
          onKeyUp={(e) => {
            if (e.key === "Escape") {
              setOpen(false);
            }
          }}
        >
          <ul className="p-0" role="none">
            {links.map((link) => (
              <li
                data-tool-nav
                key={link.href}
                className={cn(
                  "w-full text-lg text-neutral-800 first:rounded-t-lg last:rounded-b-lg",
                  {
                    "bg-neutral-700/10 font-bold": link.href === current?.href,
                    "font-semibold hover:bg-neutral-800/10":
                      link.href !== current?.href,
                  },
                )}
              >
                <a
                  href={link.href}
                  className="block h-full w-full px-2 py-2"
                  hx-get=""
                >
                  {link.label}
                </a>
              </li>
            ))}
          </ul>
        </nav>
      </Transition>
    </div>
 * */
