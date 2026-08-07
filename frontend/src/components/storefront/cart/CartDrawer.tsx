"use client";

import Link from "next/link";
import { useEffect } from "react";

import { useCart } from "@/components/storefront/cart/CartProvider";
import { formatMoney, formatQuantity } from "@/lib/format";

export function CartDrawer() {
  const { items, drawerOpen, closeDrawer, feedback, clearFeedback, hydrated, removeItem } =
    useCart();

  useEffect(() => {
    if (!drawerOpen) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") closeDrawer();
    };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [drawerOpen, closeDrawer]);

  if (!drawerOpen) return null;

  const subtotal = items.reduce(
    (sum, item) => sum + item.effectiveSalePrice * item.quantity,
    0,
  );

  return (
    <div className="fixed inset-0 z-[60]" role="dialog" aria-modal="true">
      <button
        type="button"
        className="absolute inset-0 bg-stone-950/45"
        aria-label="Zatvori korpu"
        onClick={closeDrawer}
      />
      <div className="absolute inset-y-0 right-0 flex w-full max-w-md flex-col border-l border-stone-200 bg-[#f6f4f1] shadow-[0_24px_60px_rgba(28,25,23,0.18)]">
        <div className="flex items-center justify-between border-b border-stone-200 px-5 py-4">
          <div>
            <p className="text-[11px] uppercase tracking-[0.18em] text-[#8a6a45]">
              Korpa
            </p>
            <h2 className="mt-1 font-[family-name:var(--font-storefront-display)] text-2xl text-stone-900">
              Vaša korpa
            </h2>
          </div>
          <button
            type="button"
            onClick={closeDrawer}
            className="rounded-full border border-stone-300 bg-white px-3 py-1.5 text-sm text-stone-700 transition hover:border-stone-400"
          >
            Zatvori
          </button>
        </div>

        {feedback ? (
          <div className="mx-5 mt-4 rounded-xl border border-stone-200 bg-white px-4 py-3 text-sm text-stone-700">
            <p>{feedback}</p>
            <div className="mt-2 flex flex-wrap gap-3 text-sm">
              <button
                type="button"
                className="text-stone-500 underline-offset-4 hover:text-stone-800 hover:underline"
                onClick={() => {
                  clearFeedback();
                  closeDrawer();
                }}
              >
                Nastavi kupovinu
              </button>
              <Link
                href="/korpa"
                onClick={closeDrawer}
                className="font-medium text-stone-900 underline-offset-4 hover:underline"
              >
                Pogledaj korpu
              </Link>
            </div>
          </div>
        ) : null}

        <div className="flex-1 overflow-y-auto px-5 py-4">
          {!hydrated ? (
            <p className="text-sm text-stone-500">Učitavanje…</p>
          ) : items.length === 0 ? (
            <p className="text-sm text-stone-500">Vaša korpa je prazna.</p>
          ) : (
            <ul className="space-y-4">
              {items.map((item) => (
                <li
                  key={item.productId}
                  className="flex gap-3 border-b border-stone-200/80 pb-4"
                >
                  <div className="h-16 w-16 shrink-0 overflow-hidden rounded-lg border border-stone-200 bg-white">
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
                    <div className="flex items-start justify-between gap-2">
                      <Link
                        href={`/proizvodi/${item.slug}`}
                        onClick={closeDrawer}
                        className="line-clamp-2 text-sm font-medium text-stone-900 hover:text-[#5c4630]"
                      >
                        {item.name}
                      </Link>
                      <button
                        type="button"
                        onClick={() => removeItem(item.productId)}
                        className="shrink-0 text-xs text-stone-500 underline-offset-2 transition hover:text-stone-900 hover:underline"
                        aria-label={`Ukloni ${item.name}`}
                      >
                        Ukloni
                      </button>
                    </div>
                    <p className="mt-1 text-xs text-stone-500">
                      {formatQuantity(item.quantity)}
                      {item.unit ? ` ${item.unit}` : ""}
                    </p>
                    <p className="mt-1 text-sm tabular-nums text-stone-800">
                      {formatMoney(item.effectiveSalePrice * item.quantity)}
                    </p>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="border-t border-stone-200 bg-white/70 px-5 py-5">
          <div className="flex items-baseline justify-between gap-3">
            <span className="text-sm text-stone-500">Ukupno</span>
            <span className="font-[family-name:var(--font-storefront-display)] text-xl tabular-nums text-stone-900">
              {formatMoney(subtotal)}
            </span>
          </div>
          <Link
            href="/korpa"
            onClick={closeDrawer}
            className="mt-4 flex min-h-11 w-full items-center justify-center rounded-full bg-[#141311] px-5 text-sm font-medium text-white transition hover:bg-[#2a2420]"
          >
            Pogledaj korpu
          </Link>
        </div>
      </div>
    </div>
  );
}
