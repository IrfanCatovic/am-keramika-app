"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useId, useState } from "react";

import { useAuth } from "@/components/auth/AuthProvider";
import { usePendingOrdersCount } from "@/hooks/usePendingOrdersCount";
import { getNavItemsForRole } from "@/lib/navigation";
import { userDisplayName } from "@/lib/user-display";
import { roleLabel } from "@/types/auth";

function BrandMark({ compact = false }: { compact?: boolean }) {
  return (
    <div className={`flex items-center gap-3 ${compact ? "" : ""}`}>
      <div
        className={`flex shrink-0 items-center justify-center rounded-xl bg-stone-900 text-[#d7b896] ring-1 ring-[#c4a484]/35 ${
          compact ? "h-9 w-9 text-xs font-semibold" : "h-11 w-11 text-sm font-semibold"
        }`}
        aria-hidden
      >
        AM
      </div>
      <div className="min-w-0">
        <p
          className={`truncate font-semibold tracking-tight text-stone-50 ${
            compact ? "text-sm" : "text-base"
          }`}
        >
          AM Keramika
        </p>
      </div>
    </div>
  );
}

function MenuIcon({ open }: { open: boolean }) {
  if (open) {
    return (
      <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden>
        <path
          d="M6 6l12 12M18 6L6 18"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
        />
      </svg>
    );
  }

  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden>
      <path
        d="M4 7h16M4 12h16M4 17h16"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
      />
    </svg>
  );
}

function SettingsIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-4 w-4" aria-hidden>
      <path
        d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.7"
      />
      <path
        d="M19.4 13.1a1.6 1.6 0 0 0 .3 1.8l.1.1a1.9 1.9 0 1 1-2.7 2.7l-.1-.1a1.6 1.6 0 0 0-1.8-.3 1.6 1.6 0 0 0-1 1.5V19a1.9 1.9 0 1 1-3.8 0v-.1a1.6 1.6 0 0 0-1-1.5 1.6 1.6 0 0 0-1.8.3l-.1.1a1.9 1.9 0 1 1-2.7-2.7l.1-.1a1.6 1.6 0 0 0 .3-1.8 1.6 1.6 0 0 0-1.5-1H5a1.9 1.9 0 1 1 0-3.8h.1a1.6 1.6 0 0 0 1.5-1 1.6 1.6 0 0 0-.3-1.8l-.1-.1a1.9 1.9 0 1 1 2.7-2.7l.1.1a1.6 1.6 0 0 0 1.8.3h.1a1.6 1.6 0 0 0 1-1.5V5a1.9 1.9 0 1 1 3.8 0v.1a1.6 1.6 0 0 0 1 1.5 1.6 1.6 0 0 0 1.8-.3l.1-.1a1.9 1.9 0 1 1 2.7 2.7l-.1.1a1.6 1.6 0 0 0-.3 1.8v.1a1.6 1.6 0 0 0 1.5 1H19a1.9 1.9 0 1 1 0 3.8h-.1a1.6 1.6 0 0 0-1.5 1Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.7"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function Sidebar() {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const { count: pendingOrdersCount } = usePendingOrdersCount();
  const [open, setOpen] = useState(false);
  const panelId = useId();

  useEffect(() => {
    if (!open) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  if (!user) {
    return null;
  }

  const items = getNavItemsForRole(user.role);

  function handleLogout() {
    setOpen(false);
    logout();
    router.replace("/login");
  }

  function closeMenu() {
    setOpen(false);
  }

  const nav = (
    <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-4" aria-label="Glavna navigacija">
      {items.map((item) => {
        const active =
          pathname === item.href || pathname.startsWith(`${item.href}/`);
        const className = `flex items-center gap-2 rounded-xl px-3 py-2.5 text-sm transition ${
          active
            ? "bg-stone-800 text-white shadow-sm ring-1 ring-[#c4a484]/25"
            : "text-stone-300 hover:bg-stone-900 hover:text-white"
        } ${!item.enabled ? "cursor-not-allowed opacity-50" : ""}`;

        const showPendingBadge =
          item.href === "/orders" && pendingOrdersCount > 0;

        if (!item.enabled) {
          return (
            <span key={item.href} className={className}>
              <span className="min-w-0 truncate">{item.label}</span>
            </span>
          );
        }

        return (
          <Link
            key={item.href}
            href={item.href}
            onClick={closeMenu}
            className={className}
          >
            <span className="min-w-0 truncate">{item.label}</span>
            {showPendingBadge ? (
              <span className="ml-auto inline-flex min-w-[1.25rem] items-center justify-center rounded-md bg-[#2a2420] px-1.5 text-[10px] font-medium text-[#e8d4b8]">
                {pendingOrdersCount > 99 ? "99+" : pendingOrdersCount}
              </span>
            ) : null}
          </Link>
        );
      })}
    </nav>
  );

  const footer = (
    <div className="border-t border-stone-800 px-4 py-4">
      <div className="flex items-start gap-2">
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-stone-100">
            {userDisplayName(user)}
          </p>
          <p className="text-xs text-stone-400">{roleLabel(user.role)}</p>
        </div>
        <Link
          href="/settings"
          onClick={closeMenu}
          title="Podešavanja"
          aria-label="Podešavanja"
          className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-stone-700 text-stone-300 transition hover:bg-stone-900 hover:text-white"
        >
          <SettingsIcon />
        </Link>
      </div>
      <button
        type="button"
        onClick={handleLogout}
        className="mt-3 w-full rounded-xl bg-stone-800 px-3 py-2.5 text-sm text-stone-100 transition hover:bg-stone-700"
      >
        Odjavi se
      </button>
    </div>
  );

  return (
    <>
      <header className="sticky top-0 z-30 flex items-center justify-between gap-3 border-b border-stone-200 bg-[#faf8f5]/95 px-4 py-3 backdrop-blur lg:hidden">
        <div className="flex min-w-0 items-center gap-3">
          <div
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-stone-900 text-xs font-semibold text-[#d7b896]"
            aria-hidden
          >
            AM
          </div>
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-stone-900">
              AM Keramika
            </p>
            <p className="truncate text-xs text-stone-500">
              {roleLabel(user.role)}
            </p>
          </div>
        </div>
        <button
          type="button"
          className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-stone-300 bg-white text-stone-800 shadow-sm transition hover:bg-stone-50"
          aria-label={open ? "Zatvori meni" : "Otvori meni"}
          aria-expanded={open}
          aria-controls={panelId}
          onClick={() => setOpen((value) => !value)}
        >
          <MenuIcon open={open} />
        </button>
      </header>

      <aside
        id={panelId}
        className={`fixed inset-y-0 left-0 z-50 flex w-[min(18rem,88vw)] flex-col bg-stone-950 text-stone-100 shadow-2xl transition-transform duration-300 ease-out lg:static lg:z-auto lg:h-auto lg:w-72 lg:shrink-0 lg:translate-x-0 lg:shadow-none ${
          open ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        }`}
      >
        <div className="flex items-center justify-between gap-3 border-b border-stone-800 px-4 py-4 lg:px-5 lg:py-5">
          <BrandMark />
          <button
            type="button"
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-stone-700 text-stone-300 transition hover:bg-stone-900 hover:text-white lg:hidden"
            aria-label="Zatvori meni"
            onClick={closeMenu}
          >
            <MenuIcon open />
          </button>
        </div>
        {nav}
        {footer}
      </aside>

      {open ? (
        <button
          type="button"
          aria-label="Zatvori meni"
          className="fixed inset-0 z-40 bg-stone-950/45 backdrop-blur-[1px] lg:hidden"
          onClick={closeMenu}
        />
      ) : null}
    </>
  );
}
