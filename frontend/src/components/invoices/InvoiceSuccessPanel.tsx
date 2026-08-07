"use client";

import Link from "next/link";

import { InvoiceDocumentActions } from "@/components/invoices/InvoiceDocumentActions";
import { InvoiceStatusBadge } from "@/components/invoices/InvoiceStatusBadge";
import { formatMoney } from "@/lib/format";
import { InvoiceDetails } from "@/types/invoice";

export function InvoiceSuccessPanel({
  invoice,
  customerLabel,
  onNewSale,
  newSaleLabel = "Nova prodaja",
  extraActions,
  title = "Račun je uspješno kreiran",
}: {
  invoice: Pick<
    InvoiceDetails,
    "id" | "totalAmount" | "status" | "paidAmount" | "remainingAmount"
  >;
  customerLabel: string;
  onNewSale?: () => void;
  newSaleLabel?: string;
  extraActions?: React.ReactNode;
  title?: string;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-stone-950/40 p-0 sm:items-center sm:p-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="invoice-success-title"
        className="flex max-h-[92vh] w-full max-w-lg flex-col overflow-hidden rounded-t-2xl border border-stone-200 bg-white shadow-2xl sm:rounded-2xl"
      >
        <div className="shrink-0 border-b border-stone-100 px-4 py-4 sm:px-5">
          <p className="text-[11px] font-medium uppercase tracking-[0.14em] text-[#8a6a45]">
            Uspjeh
          </p>
          <h2
            id="invoice-success-title"
            className="mt-1 text-xl font-semibold tracking-tight text-stone-900"
          >
            {title}
          </h2>
          <p className="mt-1 text-sm text-stone-600">
            Račun #{invoice.id} · {customerLabel}
          </p>
          <div className="mt-3 flex flex-wrap items-center gap-2">
            <p className="text-lg font-semibold tabular-nums text-stone-950">
              {formatMoney(invoice.totalAmount)}
            </p>
            <InvoiceStatusBadge status={invoice.status} />
          </div>
          <p className="mt-1 text-xs text-stone-500">
            Plaćeno {formatMoney(invoice.paidAmount)} · Preostalo{" "}
            {formatMoney(invoice.remainingAmount)}
          </p>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5">
          <InvoiceDocumentActions
            invoiceId={invoice.id}
            variant="stack"
            showShare={false}
            printLabel="Štampaj"
          />
          {extraActions ? <div className="mt-3 space-y-2">{extraActions}</div> : null}
        </div>

        <div className="shrink-0 space-y-2 border-t border-stone-100 px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:px-5">
          <Link
            href={`/invoices/${invoice.id}`}
            className="inline-flex min-h-11 w-full items-center justify-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
          >
            Otvori račun
          </Link>
          {onNewSale ? (
            <button
              type="button"
              onClick={onNewSale}
              className="inline-flex min-h-11 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800"
            >
              {newSaleLabel}
            </button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
