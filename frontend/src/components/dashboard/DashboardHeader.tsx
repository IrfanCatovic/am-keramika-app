"use client";

import { roleLabel, UserRole } from "@/types/auth";

export function canViewFinance(role: UserRole): boolean {
  return role === "developer" || role === "sef" || role === "menadzer";
}

export function DashboardHeader({
  username,
  role,
  date,
}: {
  username: string;
  role: UserRole;
  date: string;
}) {
  return (
    <header className="dash-enter relative overflow-hidden rounded-2xl border border-stone-800 bg-stone-950 px-4 py-5 text-stone-100 sm:px-6 sm:py-6">
      <div className="pointer-events-none absolute inset-0 marble-veil opacity-30" />
      <div className="pointer-events-none absolute -right-8 -top-12 h-32 w-32 rounded-full bg-[#c4a484]/15 blur-3xl" />
      <div className="relative min-w-0">
        <div className="flex items-start gap-3">
          <div
            className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-stone-900 text-xs font-semibold text-[#d7b896] ring-1 ring-[#c4a484]/30 sm:h-11 sm:w-11 sm:text-sm"
            aria-hidden
          >
            AM
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[11px] font-medium uppercase tracking-[0.18em] text-[#d7b896]">
              AM Keramika
            </p>
            <h1 className="mt-1 break-words text-2xl font-semibold tracking-tight sm:text-3xl">
              Dashboard
            </h1>
          </div>
        </div>
        <div className="mt-3 flex flex-col gap-1 text-sm text-stone-300 sm:flex-row sm:flex-wrap sm:items-center sm:gap-x-2">
          <span>
            Dobrodošli, <span className="text-stone-100">{username}</span>
          </span>
          <span className="hidden text-stone-600 sm:inline">·</span>
          <span>{roleLabel(role)}</span>
          <span className="hidden text-stone-600 sm:inline">·</span>
          <span className="tabular-nums text-stone-400">{date}</span>
        </div>
      </div>
    </header>
  );
}
