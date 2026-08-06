"use client";

import { InvoiceStatus } from "@/types/invoice";

const STATUS_STYLES: Record<
  InvoiceStatus,
  { label: string; className: string }
> = {
  paid: {
    label: "Plaćeno",
    className: "bg-emerald-50 text-emerald-800 ring-emerald-200",
  },
  unpaid: {
    label: "Neplaćeno",
    className: "bg-amber-50 text-amber-900 ring-amber-200",
  },
  partially_paid: {
    label: "Djelimično",
    className: "bg-sky-50 text-sky-900 ring-sky-200",
  },
  cancelled: {
    label: "Stornirano",
    className: "bg-stone-100 text-stone-600 ring-stone-200",
  },
};

export function InvoiceStatusBadge({ status }: { status: string }) {
  const config =
    STATUS_STYLES[status as InvoiceStatus] ??
    ({
      label: status,
      className: "bg-stone-100 text-stone-700 ring-stone-200",
    } as const);

  return (
    <span
      className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${config.className}`}
    >
      {config.label}
    </span>
  );
}
