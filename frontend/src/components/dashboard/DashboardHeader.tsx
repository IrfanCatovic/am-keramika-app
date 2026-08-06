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
    <header className="dash-enter relative overflow-hidden rounded-2xl border border-stone-200 bg-stone-950 px-5 py-6 text-stone-100 sm:px-7">
      <div className="pointer-events-none absolute inset-0 marble-veil opacity-40" />
      <div className="pointer-events-none absolute -right-10 -top-16 h-40 w-40 rounded-full bg-[#c4a484]/20 blur-3xl" />
      <div className="relative">
        <p className="text-xs font-medium uppercase tracking-[0.18em] text-[#d7b896]">
          AM Keramika
        </p>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight sm:text-3xl">
          Dashboard
        </h1>
        <p className="mt-2 max-w-2xl text-sm text-stone-300">
          Dobrodošli, <span className="text-stone-100">{username}</span>
          <span className="mx-2 text-stone-600">·</span>
          {roleLabel(role)}
          <span className="mx-2 text-stone-600">·</span>
          {date}
        </p>
      </div>
    </header>
  );
}
