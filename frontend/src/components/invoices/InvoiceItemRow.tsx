"use client";

import { formatMoney } from "@/lib/format";
import { InvoiceFormLine } from "@/types/invoice";

export function InvoiceItemRow({
  line,
  error,
  onQuantityChange,
  onRemove,
}: {
  line: InvoiceFormLine;
  error?: string | null;
  onQuantityChange: (quantity: number) => void;
  onRemove: () => void;
}) {
  const previewTotal = line.salePrice * line.quantity;

  return (
    <article
      className={`rounded-2xl border p-3 sm:p-4 ${
        error
          ? "border-red-200 bg-red-50/40"
          : "border-stone-200 bg-white"
      }`}
    >
      <div className="flex gap-3">
        <div className="h-14 w-14 shrink-0 overflow-hidden rounded-xl bg-stone-100">
          {line.imageUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={line.imageUrl}
              alt=""
              className="h-full w-full object-cover"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-[10px] text-stone-400">
              N/A
            </div>
          )}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="break-words font-medium text-stone-900">
                {line.name}
              </p>
              <p className="mt-0.5 text-xs text-stone-500">
                {formatMoney(line.salePrice)} / {line.unit} · Dostupno{" "}
                {line.stockQuantity}
              </p>
            </div>
            <button
              type="button"
              onClick={onRemove}
              className="rounded-lg border border-red-200 px-2.5 py-1.5 text-xs font-medium text-red-700 hover:bg-red-50"
            >
              Ukloni
            </button>
          </div>

          <div className="mt-3 flex flex-wrap items-end gap-3">
            <div>
              <label className="mb-1 block text-xs font-medium text-stone-600">
                Količina ({line.unit})
              </label>
              <input
                type="number"
                inputMode="decimal"
                step="0.01"
                min="0.01"
                max={line.stockQuantity}
                value={Number.isFinite(line.quantity) ? line.quantity : ""}
                onChange={(event) => {
                  const next = Number(event.target.value);
                  onQuantityChange(next);
                }}
                className="w-28 rounded-xl border border-stone-200 px-3 py-2 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
              />
            </div>
            <p className="pb-2 text-sm font-semibold tabular-nums text-stone-900">
              {formatMoney(previewTotal)}
            </p>
          </div>
          {error ? (
            <p className="mt-2 break-words text-xs text-red-700">{error}</p>
          ) : null}
        </div>
      </div>
    </article>
  );
}
