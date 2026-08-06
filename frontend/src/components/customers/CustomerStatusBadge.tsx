"use client";

export function CustomerStatusBadge({
  mode,
}: {
  mode: "active-list" | "all-list" | "unknown";
}) {
  if (mode === "active-list") {
    return (
      <span className="inline-flex items-center rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-800 ring-1 ring-inset ring-emerald-200">
        Aktivni pregled
      </span>
    );
  }
  if (mode === "all-list") {
    return (
      <span className="inline-flex items-center rounded-md bg-stone-100 px-2 py-0.5 text-xs font-medium text-stone-600 ring-1 ring-inset ring-stone-200">
        Uključeni i neaktivni
      </span>
    );
  }
  return (
    <span className="inline-flex items-center rounded-md bg-stone-100 px-2 py-0.5 text-xs font-medium text-stone-600 ring-1 ring-inset ring-stone-200">
      Status nije u API odgovoru
    </span>
  );
}

export function DebtBadge({ amount }: { amount: number }) {
  if (amount <= 0) {
    return (
      <span className="inline-flex items-center rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-800 ring-1 ring-inset ring-emerald-200">
        Bez duga
      </span>
    );
  }
  return (
    <span className="inline-flex items-center rounded-md bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-900 ring-1 ring-inset ring-amber-200">
      Dugovanje
    </span>
  );
}
