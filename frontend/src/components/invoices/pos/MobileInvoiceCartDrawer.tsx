"use client";

import { useEffect } from "react";

import { InvoiceCart } from "@/components/invoices/pos/InvoiceCart";
import { formatMoney, formatQuantity } from "@/lib/format";
import { InvoiceFormLine } from "@/types/invoice";

export function MobileInvoiceCartDrawer({
  open,
  onClose,
  lines,
  lineErrors,
  highlightedProductID,
  customerLabel,
  isCashSale,
  submitting,
  error,
  canSubmit,
  onQuantityChange,
  onRemove,
  onSubmitPrint,
  onSubmitNoPrint,
}: {
  open: boolean;
  onClose: () => void;
  lines: InvoiceFormLine[];
  lineErrors: Record<number, string>;
  highlightedProductID?: number | null;
  customerLabel: string;
  isCashSale: boolean;
  submitting: boolean;
  error: string | null;
  canSubmit: boolean;
  onQuantityChange: (productID: number, quantity: number) => void;
  onRemove: (productID: number) => void;
  onSubmitPrint: () => void;
  onSubmitNoPrint: () => void;
}) {
  useEffect(() => {
    if (!open) {
      return;
    }
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = previous;
      window.removeEventListener("keydown", onKey);
    };
  }, [open, onClose]);

  if (!open) {
    return null;
  }

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

  const primaryLabel = isCashSale
    ? "Naplati i štampaj"
    : "Kreiraj račun i štampaj";
  const secondaryLabel = isCashSale
    ? "Naplati bez štampe"
    : "Kreiraj bez štampe";

  return (
    <div className="fixed inset-0 z-40 lg:hidden">
      <button
        type="button"
        aria-label="Zatvori pregled računa"
        className="absolute inset-0 bg-stone-950/40"
        onClick={onClose}
      />
      <div className="absolute inset-x-0 bottom-0 flex max-h-[88vh] flex-col rounded-t-2xl border border-stone-200 bg-white shadow-2xl">
        <div className="flex items-center justify-between gap-3 border-b border-stone-100 px-4 py-3">
          <div className="min-w-0">
            <p className="text-base font-semibold text-stone-900">
              Trenutni račun
            </p>
            <p className="truncate text-sm text-stone-500">{customerLabel}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-stone-200 text-stone-600"
            aria-label="Zatvori"
          >
            ✕
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3">
          <InvoiceCart
            lines={lines}
            lineErrors={lineErrors}
            highlightedProductID={highlightedProductID}
            onQuantityChange={onQuantityChange}
            onRemove={onRemove}
          />
        </div>

        <div className="shrink-0 border-t border-stone-100 px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
          <div className="mb-2 flex justify-between text-sm">
            <span className="text-stone-500">
              {lines.length} proiz. · {formatQuantity(quantityTotal)}
            </span>
            <span className="font-semibold tabular-nums text-stone-900">
              {formatMoney(previewTotal)}
            </span>
          </div>
          {error ? (
            <p className="mb-2 break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}
          <button
            type="button"
            disabled={!canSubmit || submitting}
            onClick={onSubmitPrint}
            className="inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white disabled:opacity-50"
          >
            {submitting ? "Obrada…" : primaryLabel}
          </button>
          <button
            type="button"
            disabled={!canSubmit || submitting}
            onClick={onSubmitNoPrint}
            className="mt-2 inline-flex min-h-10 w-full items-center justify-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-600 disabled:opacity-50"
          >
            {secondaryLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
