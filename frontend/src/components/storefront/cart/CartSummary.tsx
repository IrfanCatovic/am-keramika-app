"use client";

import { formatMoney } from "@/lib/format";

export function CartSummary({
  productCount,
  subtotal,
  savings,
}: {
  productCount: number;
  subtotal: number;
  savings: number;
}) {
  return (
    <aside className="rounded-xl border border-stone-200 bg-white p-6 lg:sticky lg:top-24">
      <p className="text-[11px] uppercase tracking-[0.18em] text-[#8a6a45]">
        Pregled
      </p>
      <h2 className="mt-2 font-[family-name:var(--font-storefront-display)] text-2xl text-stone-900">
        Pregled narudžbine
      </h2>

      <dl className="mt-6 space-y-3 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Broj proizvoda</dt>
          <dd className="tabular-nums text-stone-900">{productCount}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Međuzbir</dt>
          <dd className="tabular-nums text-stone-900">
            {formatMoney(subtotal)}
          </dd>
        </div>
        {savings > 0 ? (
          <div className="flex justify-between gap-3">
            <dt className="text-stone-500">Ušteda na akcijama</dt>
            <dd className="tabular-nums text-[#5c4630]">
              −{formatMoney(savings)}
            </dd>
          </div>
        ) : null}
        <div className="flex justify-between gap-3 border-t border-stone-200 pt-3">
          <dt className="font-medium text-stone-900">Ukupno</dt>
          <dd className="font-[family-name:var(--font-storefront-display)] text-xl tabular-nums text-stone-900">
            {formatMoney(subtotal)}
          </dd>
        </div>
      </dl>

      <div className="mt-6 space-y-3 rounded-xl border border-stone-200/80 bg-[#f6f4f1] px-4 py-4 text-sm leading-relaxed text-stone-600">
        <p>Troškovi transporta nisu uračunati u cenu.</p>
        <p>
          Nakon prijema narudžbine kontaktiraćemo vas u najkraćem roku radi
          dogovora o načinu i ceni dostave.
        </p>
      </div>
      {/* Checkout CTA reserved for KORAK 4 */}
    </aside>
  );
}
