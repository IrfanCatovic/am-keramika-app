"use client";

import Link from "next/link";

import { InvoiceStatusBadge } from "@/components/invoices/InvoiceStatusBadge";
import { formatMoney } from "@/lib/format";
import { invoiceCustomerLabel } from "@/lib/invoices-api";
import { InvoiceListItem } from "@/types/invoice";

export function InvoiceCard({ invoice }: { invoice: InvoiceListItem }) {
  return (
    <article className="dash-enter min-w-0 rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <Link
            href={`/invoices/${invoice.id}`}
            className="text-base font-semibold text-stone-900 hover:text-[#8a6a45]"
          >
            Račun #{invoice.id}
          </Link>
          <p className="mt-1 break-words text-sm text-stone-600">
            {invoiceCustomerLabel(invoice)}
          </p>
          <p className="mt-1 text-xs text-stone-400">{invoice.createdAt}</p>
        </div>
        <InvoiceStatusBadge status={invoice.status} />
      </div>

      <div className="mt-3 space-y-1 text-sm">
        <p className="font-semibold text-stone-900">
          {formatMoney(invoice.totalAmount)}
        </p>
        <p className="text-stone-500">
          Plaćeno {formatMoney(invoice.paidAmount)} · Preostalo{" "}
          {formatMoney(invoice.remainingAmount)}
        </p>
        {invoice.createdByUser?.username ? (
          <p className="text-xs text-stone-400">
            Kreirao: {invoice.createdByUser.username}
          </p>
        ) : null}
      </div>

      <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
        <Link
          href={`/invoices/${invoice.id}`}
          className="inline-flex min-h-10 items-center justify-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Detalji
        </Link>
        <a
          href={`/invoices/${invoice.id}/print?autoprint=1`}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex min-h-10 items-center justify-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Štampaj
        </a>
      </div>
    </article>
  );
}
