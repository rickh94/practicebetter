import { type ReactNode, useEffect, useRef } from "preact/compat";
import { cn } from "../common";
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
    if (ref.current) {
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
  children: ReactNode;
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
  children: ReactNode;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "focusable action-button amber",
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
  children: ReactNode;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "focusable action-button green",
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
      className="focusable action-button indigo"
    >
      <span
        className="icon-[custom--music-file] -mr-1 size-6"
        aria-hidden="true"
      />
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
      className={cn("focusable action-button violet", grow && "flex-grow")}
    >
      <span
        className="icon-[custom--music-file-curly] -ml-1 size-5"
        aria-hidden="true"
      />
      Back to Plan
    </Link>
  );
}
