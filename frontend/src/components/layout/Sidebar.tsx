"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";

import { useAuth } from "@/components/auth/AuthProvider";
import { getNavItemsForRole } from "@/lib/navigation";
import { roleLabel } from "@/types/auth";

export function Sidebar() {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);

  if (!user) {
    return null;
  }

  const items = getNavItemsForRole(user.role);

  function handleLogout() {
    logout();
    router.replace("/login");
  }

  return (
    <>
      <div className="flex items-center justify-between border-b border-slate-200 bg-white px-4 py-3 lg:hidden">
        <div>
          <p className="text-sm font-semibold text-slate-900">AM Keramika</p>
          <p className="text-xs text-slate-500">{roleLabel(user.role)}</p>
        </div>
        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          className="rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700"
        >
          {open ? "Zatvori" : "Meni"}
        </button>
      </div>

      <aside
        className={`fixed inset-y-0 left-0 z-40 flex w-72 flex-col border-r border-slate-200 bg-slate-950 text-slate-100 transition-transform lg:static lg:translate-x-0 ${
          open ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="border-b border-slate-800 px-5 py-5">
          <p className="text-lg font-semibold tracking-tight">AM Keramika</p>
          <p className="mt-1 text-sm text-slate-400">Interna aplikacija</p>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-4">
          {items.map((item) => {
            const active =
              pathname === item.href || pathname.startsWith(`${item.href}/`);
            const className = `block rounded-md px-3 py-2 text-sm transition ${
              active
                ? "bg-slate-800 text-white"
                : "text-slate-300 hover:bg-slate-900 hover:text-white"
            } ${!item.enabled ? "cursor-not-allowed opacity-50" : ""}`;

            if (!item.enabled) {
              return (
                <span key={item.href} className={className}>
                  {item.label}
                </span>
              );
            }

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setOpen(false)}
                className={className}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="border-t border-slate-800 px-4 py-4">
          <p className="truncate text-sm font-medium">{user.username}</p>
          <p className="text-xs text-slate-400">{roleLabel(user.role)}</p>
          <button
            type="button"
            onClick={handleLogout}
            className="mt-3 w-full rounded-md bg-slate-800 px-3 py-2 text-sm text-slate-100 hover:bg-slate-700"
          >
            Odjavi se
          </button>
        </div>
      </aside>

      {open ? (
        <button
          type="button"
          aria-label="Zatvori meni"
          className="fixed inset-0 z-30 bg-slate-950/40 lg:hidden"
          onClick={() => setOpen(false)}
        />
      ) : null}
    </>
  );
}
