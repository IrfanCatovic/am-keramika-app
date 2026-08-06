"use client";

import { formatMoney, formatQuantity } from "@/lib/format";
import { InvoiceFormLine } from "@/types/invoice";

export function InvoiceCartItem({
  line,
  error,
  highlighted,
  onQuantityChange,
  onRemove,
}: {
  line: InvoiceFormLine;
  error?: string | null;
  highlighted?: boolean;
  onQuantityChange: (quantity: number) => void;
  onRemove: () => void;
}) {
  const previewTotal = line.salePrice * line.quantity;
  /** Minus korak −1: na količini 1 ostaje disabled (uklanjanje ide preko kante). */
  const canDecrease = Math.round((line.quantity - 1) * 100) / 100 >= 0.01;

  function bump(delta: number) {
    const next = Math.round((line.quantity + delta) * 100) / 100;
    if (next < 0.01) {
      return;
    }
    if (next > line.stockQuantity) {
      onQuantityChange(line.stockQuantity);
      return;
    }
    onQuantityChange(next);
  }

  return (
    <article
      className={`rounded-xl border p-3 transition ${
        error
          ? "border-red-200 bg-red-50/50"
          : highlighted
            ? "border-[#c4a484] bg-[#f8f1e8] ring-1 ring-[#c4a484]/40"
            : "border-stone-200 bg-white"
      }`}
    >
      <div className="flex gap-2.5">
        <div className="h-12 w-12 shrink-0 overflow-hidden rounded-lg bg-stone-100">
          {line.imageUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={line.imageUrl}
              alt=""
              className="h-full w-full object-cover"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-[9px] text-stone-400">
              N/A
            </div>
          )}
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="break-words text-sm font-medium text-stone-900">
                {line.name}
              </p>
              <p className="mt-0.5 text-[11px] text-stone-500">
                {formatMoney(line.salePrice)} / {line.unit}
              </p>
            </div>
            <button
              type="button"
              onClick={onRemove}
              aria-label={`Ukloni ${line.name}`}
              className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-stone-200 text-stone-500 transition hover:border-red-200 hover:bg-red-50 hover:text-red-700"
            >
              <svg viewBox="0 0 24 24" className="h-4 w-4" aria-hidden>
                <path
                  d="M9 3h6l1 2h4v2H4V5h4l1-2zm1 6h2v9h-2V9zm4 0h2v9h-2V9zM7 9h2v9H7V9z"
                  fill="currentColor"
                />
              </svg>
            </button>
          </div>

          <div className="mt-2.5 flex flex-wrap items-center justify-between gap-2">
            <div className="inline-flex items-center rounded-lg border border-stone-200 bg-stone-50">
              <button
                type="button"
                disabled={!canDecrease}
                onClick={() => bump(-1)}
                aria-label="Smanji količinu"
                className="inline-flex h-9 w-9 items-center justify-center text-stone-700 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
              >
                −
              </button>
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
                aria-label={`Količina ${line.name}`}
                className="h-9 w-16 border-x border-stone-200 bg-white text-center text-sm tabular-nums outline-none"
              />
              <button
                type="button"
                disabled={line.quantity >= line.stockQuantity}
                onClick={() => bump(1)}
                aria-label="Povećaj količinu"
                className="inline-flex h-9 w-9 items-center justify-center text-stone-700 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
              >
                +
              </button>
            </div>
            <p className="text-sm font-semibold tabular-nums text-stone-900">
              {formatMoney(previewTotal)}
            </p>
          </div>

          <p className="mt-1 text-[11px] text-stone-400">
            Max {formatQuantity(line.stockQuantity)} {line.unit}
          </p>

          {error ? (
            <p className="mt-1.5 break-words text-xs text-red-700">{error}</p>
          ) : null}
        </div>
      </div>
    </article>
  );
}
