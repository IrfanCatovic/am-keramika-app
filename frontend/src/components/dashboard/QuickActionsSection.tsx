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
      <div className="grid grid-cols-1 gap-3 min-[400px]:grid-cols-2 lg:grid-cols-3">
        {ACTIONS.map((action) => (
          <Link
            key={action.href}
            href={action.href}
            className="group min-w-0 rounded-2xl border border-stone-200 bg-white px-4 py-4 transition hover:-translate-y-0.5 hover:border-[#c4a484]/55 hover:shadow-[0_8px_24px_rgba(28,25,23,0.06)]"
          >
            <p className="text-sm font-semibold text-stone-900 group-hover:text-[#7a5a38]">
              {action.title}
            </p>
            <p className="mt-1 text-xs leading-relaxed text-stone-500">
              {action.description}
            </p>
          </Link>
        ))}
      </div>
    </SectionCard>
  );
}
