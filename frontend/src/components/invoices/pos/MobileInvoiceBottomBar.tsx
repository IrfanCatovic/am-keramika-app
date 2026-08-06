"use client";

import { formatMoney } from "@/lib/format";

export function MobileInvoiceBottomBar({
  itemCount,
  previewTotal,
  onReview,
}: {
  itemCount: number;
  previewTotal: number;
  onReview: () => void;
}) {
  if (itemCount === 0) {
    return null;
  }

  return (
    <div className="fixed inset-x-0 bottom-0 z-30 border-t border-stone-200 bg-white/95 px-4 py-3 backdrop-blur lg:hidden">
      <div className="mx-auto flex max-w-6xl items-center gap-3">
        <div className="min-w-0 flex-1">
          <p className="text-xs text-stone-500">
            {itemCount} {itemCount === 1 ? "stavka" : "stavke"}
          </p>
          <p className="truncate text-base font-semibold tabular-nums text-stone-900">
            {formatMoney(previewTotal)}
          </p>
        </div>
        <button
          type="button"
          onClick={onReview}
          className="inline-flex min-h-11 shrink-0 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white"
        >
          Pregled računa
        </button>
      </div>
    </div>
  );
}
