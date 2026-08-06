"use client";

import Link from "next/link";

import { SectionCard } from "@/components/dashboard/SectionCard";

const ACTIONS = [
  {
    href: "/invoices/new",
    title: "Novi račun",
    description: "Kreiraj prodaju ili zaduženje",
  },
  {
    href: "/products/new",
    title: "Dodaj proizvod",
    description: "Unesi novi artikal u katalog",
  },
  {
    href: "/customers/new",
    title: "Dodaj kupca",
    description: "Registruj novog partnera",
  },
] as const;

export function QuickActionsSection() {
  return (
    <SectionCard
      title="Brze akcije"
      description="Najčešći zadaci u radnom danu"
    >
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {ACTIONS.map((action) => (
          <Link
            key={action.href}
            href={action.href}
            className="group rounded-2xl border border-stone-200 bg-stone-50/60 px-4 py-4 transition hover:-translate-y-0.5 hover:border-[#c4a484]/50 hover:bg-white hover:shadow-md"
          >
            <p className="text-sm font-semibold text-stone-900 group-hover:text-[#7a5a38]">
              {action.title}
            </p>
            <p className="mt-1 text-xs text-stone-500">{action.description}</p>
          </Link>
        ))}
      </div>
    </SectionCard>
  );
}
