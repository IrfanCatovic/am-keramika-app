"use client";

import Image from "next/image";

import {
  COMPANY_LOGO_SRC,
  companyAddressLines,
  companyConfig,
  companyContactLines,
  companyIdLines,
} from "@/config/company";
import { formatMoney, formatQuantity } from "@/lib/format";
import { invoiceCustomerLabel } from "@/lib/invoices-api";
import { userDisplayName } from "@/lib/user-display";
import { InvoiceDetails } from "@/types/invoice";

function printStatusLabel(status: string): string {
  switch (status) {
    case "paid":
      return "Plaćen";
    case "unpaid":
      return "Neplaćen";
    case "partially_paid":
      return "Delimično plaćen";
    case "cancelled":
      return "Storniran";
    default:
      return status;
  }
}

export function InvoicePrintDocument({
  invoice,
  printedAt,
}: {
  invoice: InvoiceDetails;
  printedAt?: string;
}) {
  const isCancelled = invoice.status === "cancelled";
  const isCash = invoice.customerID == null && !invoice.customer;
  const addressLines = companyAddressLines();
  const contactLines = companyContactLines();
  const idLines = companyIdLines();
  const remainingSettled =
    !isCancelled && invoice.remainingAmount <= 0.000_001;

  return (
    <article className="invoice-print-sheet relative mx-auto bg-white text-stone-900">
      {isCancelled ? (
        <div className="invoice-print-watermark" aria-hidden>
          OTKAZANO
        </div>
      ) : null}

      <header className="invoice-print-header flex items-start justify-between gap-6 border-b border-stone-300 pb-4">
        <div className="flex min-w-0 flex-1 items-start gap-3">
          <div className="invoice-print-logo-wrap shrink-0">
            <Image
              src={COMPANY_LOGO_SRC}
              alt={companyConfig.name}
              width={220}
              height={70}
              className="invoice-print-logo h-auto w-auto max-h-[64px] object-contain object-left"
              priority
              unoptimized
            />
          </div>
          <div className="min-w-0 pt-0.5">
            <p className="text-sm font-semibold tracking-tight text-stone-900">
              {companyConfig.name}
            </p>
            {addressLines.map((line) => (
              <p key={line} className="text-xs text-stone-600">
                {line}
              </p>
            ))}
            {contactLines.map((line) => (
              <p key={line} className="text-xs text-stone-600">
                {line}
              </p>
            ))}
            {idLines.map((line) => (
              <p key={line} className="text-xs text-stone-600">
                {line}
              </p>
            ))}
          </div>
        </div>

        <div className="shrink-0 text-right">
          <h1 className="text-2xl font-bold tracking-[0.08em] text-stone-950">
            RAČUN
          </h1>
          <p className="mt-2 text-sm font-semibold text-stone-900">
            Br. {invoice.id}
          </p>
          <p className="mt-1 text-xs text-stone-600">{invoice.createdAt}</p>
          <p
            className={`mt-2 inline-block rounded border px-2 py-0.5 text-xs font-semibold uppercase tracking-wide ${
              isCancelled
                ? "border-stone-400 text-stone-700"
                : "border-stone-300 text-stone-800"
            }`}
          >
            {printStatusLabel(invoice.status)}
          </p>
        </div>
      </header>

      <section className="mt-5">
        <h2 className="text-[11px] font-semibold uppercase tracking-[0.14em] text-stone-500">
          Kupac
        </h2>
        {isCash ? (
          <p className="mt-1 text-sm font-medium text-stone-900">
            Gotovinska prodaja
          </p>
        ) : (
          <div className="mt-1">
            <p className="text-sm font-medium text-stone-900">
              {invoiceCustomerLabel(invoice)}
            </p>
            {invoice.customer?.phone ? (
              <p className="text-xs text-stone-600">{invoice.customer.phone}</p>
            ) : null}
          </div>
        )}
      </section>

      <section className="mt-5">
        <table className="invoice-print-table w-full border-collapse text-xs">
          <thead>
            <tr className="border-b border-stone-400 text-left">
              <th className="w-8 py-2 pr-2 font-semibold">R.br.</th>
              <th className="py-2 pr-2 font-semibold">Proizvod</th>
              <th className="w-16 py-2 pr-2 text-right font-semibold">
                Količina
              </th>
              <th className="w-14 py-2 pr-2 font-semibold">Jedinica</th>
              <th className="w-24 py-2 pr-2 text-right font-semibold">
                Jedinična cijena
              </th>
              <th className="w-24 py-2 text-right font-semibold">Ukupno</th>
            </tr>
          </thead>
          <tbody>
            {(invoice.items ?? []).map((item, index) => (
              <tr
                key={`${item.productID}-${index}`}
                className="border-b border-stone-200 align-top"
              >
                <td className="py-2 pr-2 tabular-nums text-stone-600">
                  {index + 1}
                </td>
                <td className="py-2 pr-2 font-medium text-stone-900">
                  {item.productName || `Proizvod #${item.productID}`}
                </td>
                <td className="py-2 pr-2 text-right tabular-nums">
                  {formatQuantity(item.quantity)}
                </td>
                <td className="py-2 pr-2 text-stone-700">{item.unit || "—"}</td>
                <td className="py-2 pr-2 text-right tabular-nums">
                  {formatMoney(item.unitPrice)}
                </td>
                <td className="py-2 text-right tabular-nums font-medium">
                  {formatMoney(item.totalPrice)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {(invoice.items ?? []).length === 0 ? (
          <p className="mt-3 text-xs text-stone-500">Nema stavki.</p>
        ) : null}
      </section>

      <section className="mt-6 flex justify-end">
        <dl className="w-full max-w-xs space-y-1.5 text-sm">
          <div className="flex justify-between gap-6">
            <dt className="text-stone-600">Ukupno</dt>
            <dd className="font-semibold tabular-nums">
              {formatMoney(invoice.totalAmount)}
            </dd>
          </div>
          <div className="flex justify-between gap-6">
            <dt className="text-stone-600">Plaćeno</dt>
            <dd className="tabular-nums">{formatMoney(invoice.paidAmount)}</dd>
          </div>
          <div className="flex justify-between gap-6 border-t border-stone-300 pt-2">
            <dt className="font-medium text-stone-800">
              {remainingSettled ? "Status plaćanja" : "Preostalo"}
            </dt>
            <dd className="font-semibold tabular-nums">
              {isCancelled
                ? printStatusLabel("cancelled")
                : remainingSettled
                  ? "Plaćeno"
                  : formatMoney(invoice.remainingAmount)}
            </dd>
          </div>
        </dl>
      </section>

      <footer className="mt-10 grid grid-cols-2 gap-8 border-t border-stone-200 pt-6 text-xs text-stone-600">
        <div>
          {invoice.createdByUser ? (
            <p>
              Račun kreirao:{" "}
              <span className="font-medium text-stone-800">
                {userDisplayName(invoice.createdByUser)}
              </span>
            </p>
          ) : null}
          {printedAt ? (
            <p className="mt-1">Dokument odštampan: {printedAt}</p>
          ) : null}
          <div className="mt-10">
            <div className="h-10 border-b border-stone-400" />
            <p className="mt-1">Potpis kupca</p>
          </div>
        </div>
        <div className="text-right">
          <div className="mt-[4.5rem]">
            <div className="ml-auto h-10 w-48 border-b border-stone-400" />
            <p className="mt-1">Potpis / pečat prodavca</p>
          </div>
        </div>
      </footer>
    </article>
  );
}
