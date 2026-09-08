"use client";

import { formatMoney, formatQuantity, formatUnit } from "@/lib/format";
import {
  calculatePackagePrice,
  getActualProductQuantity,
  isValidInvoicePriceOverride,
  resolveInvoiceFormLineUnitPrice,
} from "@/lib/product-pricing";
import { InvoiceFormLine } from "@/types/invoice";

export function InvoiceCartItem({
  line,
  error,
  highlighted,
  onQuantityChange,
  onPriceOverrideChange,
  onRemove,
}: {
  line: InvoiceFormLine;
  error?: string | null;
  highlighted?: boolean;
  onQuantityChange: (quantity: number) => void;
  onPriceOverrideChange: (enabled: boolean, priceOverride: number | null) => void;
  onRemove: () => void;
}) {
  const quantityDetails = getActualProductQuantity(
    line.quantity,
    line.saleByPackage,
    line.packageQuantity,
  );
  const overrideEnabled = Boolean(line.priceOverrideEnabled);
  const unitPrice = resolveInvoiceFormLineUnitPrice(line);
  const previewTotal = unitPrice * quantityDetails.actualQuantity;
  const unitLabel = formatUnit(line.unit);
  /** Minus korak −1: na količini 1 ostaje disabled (uklanjanje ide preko kante). */
  const canDecrease = Math.round((line.quantity - 1) * 100) / 100 >= 0.01;
  const overrideInputValue =
    line.priceOverride == null || !Number.isFinite(line.priceOverride)
      ? ""
      : line.priceOverride;
  const showOverridePriceHint =
    overrideEnabled && isValidInvoicePriceOverride(line.priceOverride);

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
            : overrideEnabled
              ? "border-[#c4a484]/55 bg-[#faf7f3]"
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
              <div className="flex flex-wrap items-center gap-1.5">
                <p className="break-words text-sm font-medium text-stone-900">
                  {line.name}
                </p>
                {overrideEnabled ? (
                  <span className="inline-flex rounded-md bg-[#2a2420]/90 px-1.5 py-0.5 text-[10px] font-medium tracking-wide text-[#e8d5bc]">
                    Ručna cena
                  </span>
                ) : null}
              </div>
              <p className="mt-0.5 text-[11px] text-stone-500">
                {overrideEnabled
                  ? `Redovna cena: ${formatMoney(line.salePrice)} / ${unitLabel}`
                  : `${formatMoney(line.salePrice)} / ${unitLabel}`}
              </p>
              {line.saleByPackage && line.packageQuantity ? (
                <div className="mt-1 text-xs text-stone-600">
                  <p>
                    Potrebno: {formatQuantity(line.quantity)} {unitLabel}
                  </p>
                  <p>
                    {quantityDetails.packageCount} paketa ×{" "}
                    {formatQuantity(line.packageQuantity)} {unitLabel}
                  </p>
                  <p>
                    Obračun: {formatQuantity(quantityDetails.actualQuantity)}{" "}
                    {unitLabel}
                  </p>
                  <p>
                    {formatMoney(unitPrice)} / {unitLabel} ·{" "}
                    {formatMoney(
                      calculatePackagePrice(unitPrice, line.packageQuantity),
                    )}{" "}
                    / paket
                  </p>
                </div>
              ) : null}
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
                aria-label={`${line.saleByPackage ? "Potrebna količina" : "Količina"} ${line.name}`}
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

          <label className="mt-2.5 flex cursor-pointer items-center gap-2 text-xs text-stone-700">
            <input
              type="checkbox"
              checked={overrideEnabled}
              onChange={(event) => {
                if (event.target.checked) {
                  onPriceOverrideChange(true, line.priceOverride ?? null);
                } else {
                  onPriceOverrideChange(false, null);
                }
              }}
              className="h-3.5 w-3.5 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
            />
            <span>Popust na kasi</span>
          </label>

          {overrideEnabled ? (
            <div className="mt-2 space-y-1.5">
              <label className="block text-[11px] font-medium text-stone-600">
                Cena na kasi
              </label>
              <div className="flex flex-wrap items-center gap-2">
                <input
                  type="number"
                  inputMode="decimal"
                  step="0.01"
                  min="0.01"
                  value={overrideInputValue}
                  onChange={(event) => {
                    const raw = event.target.value;
                    if (raw.trim() === "") {
                      onPriceOverrideChange(true, null);
                      return;
                    }
                    onPriceOverrideChange(true, Number(raw));
                  }}
                  aria-label={`Cena na kasi za ${line.name}`}
                  className="h-9 w-28 rounded-lg border border-stone-200 bg-white px-2.5 text-sm tabular-nums outline-none ring-[#c4a484]/35 focus:ring-2"
                />
                <span className="text-xs text-stone-500">
                  RSD / {unitLabel}
                </span>
              </div>
              {showOverridePriceHint ? (
                <div className="text-[11px] leading-relaxed text-stone-500">
                  <p>
                    Redovna cena: {formatMoney(line.salePrice)} / {unitLabel}
                  </p>
                  <p>
                    Cena na kasi: {formatMoney(unitPrice)} / {unitLabel}
                  </p>
                </div>
              ) : null}
            </div>
          ) : null}

          <p className="mt-1 text-[11px] text-stone-400">
            Max {formatQuantity(line.stockQuantity)} {unitLabel}
          </p>

          {error ? (
            <p className="mt-1.5 break-words text-xs text-red-700">{error}</p>
          ) : null}
        </div>
      </div>
    </article>
  );
}
