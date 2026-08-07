"use client";

import Link from "next/link";

import { PublicProductPrice } from "@/components/storefront/PublicPrice";
import { formatMoney, formatQuantity } from "@/lib/format";
import type { CartItem } from "@/types/cart";

export function CheckoutSummary({
  items,
  subtotal,
}: {
  items: CartItem[];
  subtotal: number;
}) {
  return (
    <aside className="rounded-xl border border-stone-200 bg-white p-6 lg:sticky lg:top-24">
      <p className="text-[11px] uppercase tracking-[0.18em] text-[#8a6a45]">
        Pregled
      </p>
      <h2 className="mt-2 font-[family-name:var(--font-storefront-display)] text-2xl text-stone-900">
        Pregled narudžbine
      </h2>

      <ul className="mt-6 space-y-4">
        {items.map((item) => (
          <li
            key={item.productId}
            className="flex gap-3 border-b border-stone-100 pb-4 last:border-0"
          >
            <div className="h-16 w-16 shrink-0 overflow-hidden rounded-lg border border-stone-200 bg-[#f7f5f2]">
              {item.imageUrl ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={item.imageUrl}
                  alt=""
                  className="h-full w-full object-contain p-1"
                />
              ) : (
                <div className="flex h-full items-center justify-center text-xs text-stone-300">
                  AM
                </div>
              )}
            </div>
            <div className="min-w-0 flex-1">
              <p className="line-clamp-2 text-sm font-medium text-stone-900">
                {item.name}
              </p>
              <p className="mt-1 text-xs text-stone-500">
                {formatQuantity(item.quantity)}
                {item.unit ? ` ${item.unit}` : ""}
              </p>
              <div className="mt-1.5 flex flex-wrap items-baseline justify-between gap-2">
                <PublicProductPrice
                  product={{
                    salePrice: item.salePrice,
                    effectiveSalePrice: item.effectiveSalePrice,
                    isOnSale: item.isOnSale,
                    discountPercent: item.discountPercent,
                  }}
                  size="sm"
                />
                <span className="text-sm tabular-nums text-stone-800">
                  {formatMoney(item.effectiveSalePrice * item.quantity)}
                </span>
              </div>
            </div>
          </li>
        ))}
      </ul>

      <dl className="mt-2 space-y-3 border-t border-stone-200 pt-4 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Proizvodi</dt>
          <dd className="tabular-nums text-stone-900">{formatMoney(subtotal)}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Transport</dt>
          <dd className="text-stone-600">Dogovor naknadno</dd>
        </div>
        <div className="flex justify-between gap-3 border-t border-stone-200 pt-3">
          <dt className="font-medium text-stone-900">Ukupno za proizvode</dt>
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
        <p className="text-stone-500">
          Plaćanje se dogovara nakon potvrde narudžbine.
        </p>
      </div>

      <Link
        href="/korpa"
        className="mt-5 inline-flex text-sm text-stone-500 underline-offset-4 hover:text-stone-800 hover:underline"
      >
        Vrati se u korpu
      </Link>
    </aside>
  );
}
