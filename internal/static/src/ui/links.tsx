import { BookOpenIcon, HomeIcon } from "@heroicons/react/20/solid";
import { UserIcon, UserMinusIcon } from "@heroicons/react/20/solid";
import { ReactNode, useEffect, useRef } from "preact/compat";
import { cn } from "../common";
import htmx from "htmx.org";

export const topNavClasses =
  "focusable flex h-14 items-center gap-2 rounded-xl bg-neutral-700/10 px-6 py-4 font-semibold text-neutral-700 transition-all duration-200 hover:bg-neutral-700/20";

export function Link({
  className = "",
  href,
  external = false,
  target = "#main-content",
  swap = "outerHTML",
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
    htmx.process(ref.current);
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

export function BackToHome() {
  return (
    <TopNavLink href="/">
      <HomeIcon className="-ml-1 h-6 w-6" /> Back Home
    </TopNavLink>
  );
}

export function LibraryLink() {
  return (
    <TopNavLink href="/library">
      <BookOpenIcon className="-ml-1 inline h-6 w-6" />
      Library
    </TopNavLink>
  );
}
export function AccountLink() {
  return (
    <TopNavLink href="/account">
      <UserIcon className="-ml-1 inline h-6 w-6" />
      <span>Account</span>
    </TopNavLink>
  );
}

export function LoginLink() {
  return <TopNavLink href="/signin">Login →</TopNavLink>;
}

export function LogoutLink() {
  return (
    <TopNavLink href="/api/auth/signout">
      Logout
      <UserMinusIcon className="-mr-1 inline h-6 w-6" />
    </TopNavLink>
  );
}

export function BackToPieceLink({ pieceHref }: { pieceHref: string }) {
  return (
    <Link
      href={pieceHref}
      className="focusable block rounded-xl bg-sky-700/10 px-4 py-2 font-semibold text-sky-800 transition duration-200 hover:bg-sky-700/20"
    >
      Back to Piece
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
        "focusable flex items-center justify-center gap-2 rounded-xl bg-yellow-700/10 px-4 py-2 font-semibold text-yellow-800 transition duration-200 hover:bg-yellow-700/20",
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
        "focusable flex items-center justify-center gap-1 rounded-xl bg-green-700/10 px-4 py-2 font-semibold text-green-800 transition duration-200 hover:bg-green-700/20",
        grow && "flex-grow",
        className,
      )}
    >
      {children}
    </Link>
  );
}
