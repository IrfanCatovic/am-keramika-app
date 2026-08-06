"use client";

import Link from "next/link";

import { InvoiceStatusBadge } from "@/components/invoices/InvoiceStatusBadge";
import { formatMoney } from "@/lib/format";
import { paymentCustomerLabel } from "@/lib/payments-api";
import { Payment } from "@/types/payment";

export function PaymentDetails({ payment }: { payment: Payment }) {
  const customerId = payment.customer?.id ?? payment.customerID ?? null;
  const singleInvoice =
    payment.allocations?.length === 1 ? payment.allocations[0] : null;

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            Uplata
          </p>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
            #{payment.id}
          </h1>
          <p className="mt-1 text-sm text-stone-500">{payment.createdAt}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link
            href="/payments"
            className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
          >
            Lista uplata
          </Link>
          {customerId ? (
            <>
              <Link
                href={`/customers/${customerId}`}
                className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
              >
                Otvori kupca
              </Link>
              <Link
                href={`/invoices/new?customerID=${customerId}`}
                className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
              >
                Novi račun za kupca
              </Link>
            </>
          ) : null}
          {singleInvoice ? (
            <Link
              href={`/invoices/${singleInvoice.invoiceID}`}
              className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800"
            >
              Otvori račun
            </Link>
          ) : null}
        </div>
      </header>

      <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
              Kupac
            </dt>
            <dd className="mt-1 text-sm font-medium text-stone-900">
              {paymentCustomerLabel(payment)}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
              Ukupan iznos
            </dt>
            <dd className="mt-1 text-lg font-semibold tabular-nums text-stone-900">
              {formatMoney(payment.totalAmount)}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
              Evidentirao
            </dt>
            <dd className="mt-1 text-sm text-stone-800">
              {payment.createdByUser?.username ??
                `Korisnik #${payment.createdByUserID}`}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
              Datum
            </dt>
            <dd className="mt-1 text-sm text-stone-800">{payment.createdAt}</dd>
          </div>
        </dl>
      </section>

      <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <h2 className="text-base font-semibold text-stone-900">Raspodjela</h2>
        <ul className="mt-3 space-y-2">
          {(payment.allocations ?? []).map((allocation) => (
            <li
              key={allocation.id}
              className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-stone-100 px-3 py-3"
            >
              <div className="min-w-0">
                <Link
                  href={`/invoices/${allocation.invoiceID}`}
                  className="font-medium text-stone-900 hover:text-[#8a6a45]"
                >
                  Račun #{allocation.invoiceID}
                </Link>
                <div className="mt-1 flex flex-wrap items-center gap-2">
                  <InvoiceStatusBadge status={allocation.invoice.status} />
                  <span className="text-xs text-stone-500">
                    Račun {formatMoney(allocation.invoice.totalAmount)} · plaćeno{" "}
                    {formatMoney(allocation.invoice.paidAmount)}
                  </span>
                </div>
              </div>
              <p className="text-sm font-semibold tabular-nums text-stone-900">
                {formatMoney(allocation.amount)}
              </p>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
