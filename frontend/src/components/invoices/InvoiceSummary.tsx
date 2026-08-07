"use client";

import { formatMoney } from "@/lib/format";
import { InvoiceFormLine } from "@/types/invoice";

export function InvoiceSummary({
  lines,
  customerMode,
  customerName,
  submitting,
  error,
  onSubmit,
}: {
  lines: InvoiceFormLine[];
  customerMode: "cash" | "customer";
  customerName: string | null;
  submitting: boolean;
  error: string | null;
  onSubmit: () => void;
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

  return (
    <aside className="rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)] sm:p-5 lg:sticky lg:top-4">
      <h2 className="text-base font-semibold text-stone-900">Pregled računa</h2>
      <p className="mt-1 text-sm text-stone-500">
        {customerMode === "cash"
          ? "Gotovinska prodaja"
          : customerName
            ? `Kupac: ${customerName}`
            : "Kupac nije izabran"}
      </p>

      <dl className="mt-4 space-y-2 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Stavke</dt>
          <dd className="font-medium text-stone-900">{itemCount}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Ukupna količina</dt>
          <dd className="font-medium tabular-nums text-stone-900">
            {quantityTotal.toFixed(2)}
          </dd>
        </div>
        <div className="flex justify-between gap-3 border-t border-stone-100 pt-3">
          <dt className="text-stone-700">Preview total</dt>
          <dd className="text-lg font-semibold tabular-nums text-stone-950">
            {formatMoney(previewTotal)}
          </dd>
        </div>
      </dl>

      <p className="mt-3 text-xs leading-relaxed text-stone-500">
        Konačan obračun i proveru cena izvršava server prilikom kreiranja
        računa.
      </p>

      {error ? (
        <p className="mt-3 break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      ) : null}

      <button
        type="button"
        disabled={submitting || itemCount === 0}
        onClick={onSubmit}
        className="mt-4 inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800 disabled:opacity-60"
      >
        {submitting ? "Kreiranje..." : "Kreiraj račun"}
      </button>
    </aside>
  );
}
