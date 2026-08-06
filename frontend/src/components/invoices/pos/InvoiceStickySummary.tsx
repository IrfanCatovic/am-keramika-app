"use client";

import type { ReactNode } from "react";

import { formatMoney, formatQuantity } from "@/lib/format";
import { InvoiceFormLine } from "@/types/invoice";

/** Sticky desni panel: lista stavki + summary + submit. */
export function InvoiceStickyCartPanel({
  lines,
  customerLabel,
  isCashSale,
  submitting,
  error,
  canSubmit,
  onSubmit,
  cart,
}: {
  lines: InvoiceFormLine[];
  customerLabel: string;
  isCashSale: boolean;
  submitting: boolean;
  error: string | null;
  canSubmit: boolean;
  onSubmit: () => void;
  cart: ReactNode;
}) {
  const itemCount = lines.length;
  const quantityTotal = lines.reduce(
    (sum, line) => sum + (Number.isFinite(line.quantity) ? line.quantity : 0),
    0,
  );
  const previewTotal = lines.reduce(
    (sum, line) =>
      sum +
      (Number.isFinite(line.quantity) ? line.salePrice * line.quantity : 0),
    0,
  );

  const primaryLabel = isCashSale ? "Naplati" : "Kreiraj račun";

  return (
    <aside className="flex max-h-[calc(100vh-5.5rem)] flex-col overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)] lg:sticky lg:top-4">
      <div className="shrink-0 border-b border-stone-100 px-4 py-3">
        <h2 className="text-base font-semibold text-stone-900">
          Trenutni račun
        </h2>
        <p className="mt-0.5 truncate text-sm text-stone-500">{customerLabel}</p>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-3 py-3">
        {cart}
      </div>

      <div className="shrink-0 border-t border-stone-100 px-4 py-3">
        <dl className="space-y-1.5 text-sm">
          <div className="flex justify-between gap-3">
            <dt className="text-stone-500">Proizvodi</dt>
            <dd className="font-medium tabular-nums text-stone-900">
              {itemCount}
            </dd>
          </div>
          <div className="flex justify-between gap-3">
            <dt className="text-stone-500">Zbir količina</dt>
            <dd className="font-medium tabular-nums text-stone-900">
              {formatQuantity(quantityTotal)}
            </dd>
          </div>
          <div className="flex justify-between gap-3 border-t border-stone-100 pt-2">
            <dt className="font-medium text-stone-700">Ukupno</dt>
            <dd className="text-lg font-semibold tabular-nums text-stone-950">
              {formatMoney(previewTotal)}
            </dd>
          </div>
        </dl>

        <p className="mt-2 text-[11px] leading-relaxed text-stone-500">
          Konačne cijene i stanje lagera provjerava server.
        </p>

        {error ? (
          <p className="mt-2 break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        ) : null}

        <button
          type="button"
          disabled={!canSubmit || submitting}
          onClick={onSubmit}
          className="mt-3 inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white transition hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitting ? "Obrada…" : primaryLabel}
        </button>
      </div>
    </aside>
  );
}
