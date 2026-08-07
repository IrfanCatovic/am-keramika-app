"use client";

import Link from "next/link";

import { formatMoney } from "@/lib/format";

export function CheckoutSuccess({
  orderId,
  totalAmount,
}: {
  orderId: number;
  totalAmount: number;
}) {
  return (
    <div className="mx-auto max-w-xl px-4 py-16 text-center sm:px-6 sm:py-20">
      <p className="text-[11px] uppercase tracking-[0.2em] text-[#8a6a45]">
        Narudžbina primljena
      </p>
      <h1 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900 sm:text-4xl">
        Hvala na narudžbini.
      </h1>
      <p className="mt-4 text-sm leading-relaxed text-stone-600 sm:text-base">
        Vaša narudžbina #{orderId} je uspešno poslata.
      </p>
      <p className="mt-3 text-sm leading-relaxed text-stone-600 sm:text-base">
        Kontaktiraćemo vas u najkraćem roku radi potvrde narudžbine i dogovora o
        transportu.
      </p>

      <div className="mx-auto mt-8 max-w-md space-y-3 rounded-xl border border-stone-200 bg-white px-5 py-5 text-left text-sm text-stone-600">
        <p>
          Vrednost proizvoda:{" "}
          <span className="font-medium tabular-nums text-stone-900">
            {formatMoney(totalAmount)}
          </span>
        </p>
        <p className="text-stone-500">
          Troškovi transporta nisu uračunati u prikazani iznos.
        </p>
      </div>

      <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:justify-center">
        <Link
          href="/proizvodi"
          className="inline-flex min-h-11 items-center justify-center rounded-full bg-[#141311] px-6 text-sm font-medium text-white transition hover:bg-[#2a2420]"
        >
          Nastavi kupovinu
        </Link>
        <Link
          href="/"
          className="inline-flex min-h-11 items-center justify-center rounded-full border border-stone-300 bg-white px-6 text-sm font-medium text-stone-800 transition hover:border-stone-400"
        >
          Početna
        </Link>
      </div>
    </div>
  );
}
