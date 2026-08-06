"use client";

export type InvoiceSaleType = "cash" | "customer";

export function InvoiceSaleTypeSwitch({
  value,
  onChange,
  disabled = false,
}: {
  value: InvoiceSaleType;
  onChange: (next: InvoiceSaleType) => void;
  disabled?: boolean;
}) {
  return (
    <div
      className="inline-flex w-full rounded-xl border border-stone-200 bg-stone-100/80 p-1 sm:w-auto"
      role="group"
      aria-label="Tip prodaje"
    >
      <button
        type="button"
        disabled={disabled}
        onClick={() => onChange("cash")}
        className={`min-h-10 flex-1 rounded-lg px-3 text-sm font-medium transition sm:flex-none sm:px-4 ${
          value === "cash"
            ? "bg-stone-900 text-white shadow-sm"
            : "text-stone-600 hover:bg-white/70 hover:text-stone-900"
        } disabled:opacity-60`}
      >
        Gotovinska prodaja
      </button>
      <button
        type="button"
        disabled={disabled}
        onClick={() => onChange("customer")}
        className={`min-h-10 flex-1 rounded-lg px-3 text-sm font-medium transition sm:flex-none sm:px-4 ${
          value === "customer"
            ? "bg-stone-900 text-white shadow-sm"
            : "text-stone-600 hover:bg-white/70 hover:text-stone-900"
        } disabled:opacity-60`}
      >
        Kupac
      </button>
    </div>
  );
}
