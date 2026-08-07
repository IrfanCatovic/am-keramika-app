"use client";

export type InvoicePaymentMode = "unpaid" | "full" | "partial";

const OPTIONS: { value: InvoicePaymentMode; label: string }[] = [
  { value: "unpaid", label: "Bez uplate" },
  { value: "full", label: "Plati sve" },
  { value: "partial", label: "Plati deo" },
];

/** Kompaktan izbor početne uplate — samo za račun za kupca. */
export function InvoicePaymentModePicker({
  mode,
  amount,
  amountError,
  disabled,
  onModeChange,
  onAmountChange,
}: {
  mode: InvoicePaymentMode;
  amount: string;
  amountError: string | null;
  disabled?: boolean;
  onModeChange: (mode: InvoicePaymentMode) => void;
  onAmountChange: (value: string) => void;
}) {
  return (
    <section className="rounded-2xl border border-stone-200 bg-white p-3 sm:p-4">
      <h2 className="text-sm font-semibold text-stone-900">Plaćanje</h2>
      <div
        className="mt-2 grid grid-cols-1 gap-2 min-[375px]:grid-cols-3"
        role="radiogroup"
        aria-label="Način plaćanja"
      >
        {OPTIONS.map((option) => {
          const selected = mode === option.value;
          return (
            <button
              key={option.value}
              type="button"
              role="radio"
              aria-checked={selected}
              disabled={disabled}
              onClick={() => onModeChange(option.value)}
              className={`min-h-10 rounded-xl border px-2 py-2 text-center text-sm font-medium transition ${
                selected
                  ? "border-stone-900 bg-stone-900 text-white"
                  : "border-stone-200 bg-stone-50 text-stone-700 hover:border-stone-300"
              } disabled:opacity-50`}
            >
              {option.label}
            </button>
          );
        })}
      </div>

      {mode === "partial" ? (
        <label className="mt-3 block">
          <span className="text-xs font-medium text-stone-600">Iznos uplate</span>
          <div className="mt-1 flex items-center gap-2">
            <input
              type="number"
              inputMode="decimal"
              min={0}
              step="0.01"
              value={amount}
              disabled={disabled}
              onChange={(event) => onAmountChange(event.target.value)}
              className="min-h-11 w-full min-w-0 rounded-xl border border-stone-200 bg-white px-3 text-sm tabular-nums text-stone-900 outline-none focus:border-stone-400"
              placeholder="npr. 20000"
            />
            <span className="shrink-0 text-sm text-stone-500">RSD</span>
          </div>
          {amountError ? (
            <p className="mt-1.5 text-sm text-red-700">{amountError}</p>
          ) : null}
        </label>
      ) : null}
    </section>
  );
}
