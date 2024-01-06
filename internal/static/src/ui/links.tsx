import { ReactNode, useEffect, useRef } from "preact/compat";
import { cn } from "../common";
import { NoteSheetIcon } from "./icons";
import * as htmx from "htmx.org";

export const topNavClasses =
  "focusable flex h-14 items-center gap-2 rounded-xl bg-neutral-700/10 px-6 py-4 font-semibold text-neutral-700 transition-all duration-200 hover:bg-neutral-700/20";

export function Link({
  className = "",
  href,
  external = false,
  target = "#main-content",
  swap = "outerHTML transition:true",
  pushUrl = true,
  children,
}: {
  className?: string;
  href: string;
  external?: boolean;
  target?: string;
  swap?: string;
  pushUrl?: boolean;
  children: ReactNode;
}) {
  const ref = useRef<HTMLAnchorElement>(null);
  useEffect(() => {
    if (htmx) {
      htmx.process(ref.current);
    }
  }, [href, external, target, swap, pushUrl, children]);

  return (
    <a
      ref={ref}
      href={href}
      hx-get={external ? "" : href}
      hx-swap={swap}
      hx-push-url={pushUrl ? "true" : "false"}
      hx-target={target}
      className={className}
    >
      {children}
    </a>
  );
}

export function TopNavLink({
  href,
  children,
}: {
  href: string;
  children: React.ReactNode;
}) {
  return (
    <Link className={topNavClasses} href={href}>
      {children}
    </Link>
  );
}

export function WarningLink({
  href,
  className = "",
  grow = false,
  children,
}: {
  href: string;
  className?: string;
  grow?: boolean;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "focusable flex items-center justify-center gap-1 rounded-xl bg-yellow-700/10 px-4 py-2 font-semibold text-yellow-800 transition duration-200 hover:bg-yellow-700/20",
        grow && "flex-grow",
        className,
      )}
    >
      {children}
    </Link>
  );
}

export function HappyLink({
  href,
  className = "",
  grow = false,
  children,
}: {
  href: string;
  className?: string;
  grow?: boolean;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "focusable action-button bg-green-700/10 text-green-800 hover:bg-green-700/20",
        grow && "flex-grow",
        className,
      )}
    >
      {children}
    </Link>
  );
}

export function BackToPiece({ pieceid }: { pieceid: string }) {
  return (
    <Link
      href={`/library/pieces/${pieceid}`}
      className="focusable action-button bg-sky-700/10 text-sky-800 hover:bg-sky-700/20"
    >
      <NoteSheetIcon className="-ml-1 size-5" />
      Back to Piece
    </Link>
  );
}

export function BackToPlan({
  planid,
  grow = false,
}: {
  planid: string;
  grow?: boolean;
}) {
  return (
    <Link
      href={`/library/plans/${planid}`}
      className={cn(
        "focusable action-button bg-violet-700/10  text-violet-800 hover:bg-violet-700/20",
        grow && "flex-grow",
      )}
    >
      <span
        className="icon-[solar--clipboard-check-bold] -mr-1 size-5"
        aria-hidden="true"
      ></span>
      Back to Plan
    </Link>
  );
}
