"use client";

import Link from "next/link";

import { InvoiceCard } from "@/components/invoices/InvoiceCard";
import { InvoiceStatusBadge } from "@/components/invoices/InvoiceStatusBadge";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { formatMoney } from "@/lib/format";
import { invoiceCustomerLabel } from "@/lib/invoices-api";
import { InvoiceListItem } from "@/types/invoice";

export function InvoiceList({
  invoices,
  loading,
  error,
  filtersActive,
  onRetry,
}: {
  invoices: InvoiceListItem[];
  loading: boolean;
  error: string | null;
  filtersActive: boolean;
  onRetry: () => void;
}) {
  if (loading) {
    return <ListSkeleton rows={5} />;
  }

  if (error) {
    return <InlineError message={error} onRetry={onRetry} />;
  }

  if (invoices.length === 0) {
    return (
      <EmptyState
        title={filtersActive ? "Nema rezultata" : "Nema računa"}
        description={
          filtersActive
            ? "Pokušajte drugačije filtere ili pretragu."
            : "Kreirajte prvi račun da biste pratili prodaju."
        }
        action={
          !filtersActive ? (
            <Link
              href="/invoices/new"
              className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
            >
              Novi račun
            </Link>
          ) : undefined
        }
      />
    );
  }

  return (
    <>
      <ul className="space-y-3 lg:hidden">
        {invoices.map((invoice) => (
          <li key={invoice.id}>
            <InvoiceCard invoice={invoice} />
          </li>
        ))}
      </ul>

      <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
        <table className="w-full table-fixed text-left text-sm">
          <thead className="sticky top-0 bg-stone-50/95 backdrop-blur">
            <tr className="border-b border-stone-200 text-xs uppercase tracking-[0.08em] text-stone-500">
              <th className="w-[10%] px-4 py-3 font-medium">Račun</th>
              <th className="w-[22%] px-4 py-3 font-medium">Kupac</th>
              <th className="w-[12%] px-4 py-3 font-medium text-right">Ukupno</th>
              <th className="w-[12%] px-4 py-3 font-medium text-right">
                Plaćeno
              </th>
              <th className="w-[12%] px-4 py-3 font-medium text-right">
                Preostalo
              </th>
              <th className="w-[12%] px-4 py-3 font-medium">Status</th>
              <th className="w-[12%] px-4 py-3 font-medium">Datum</th>
              <th className="px-4 py-3 font-medium">Akcije</th>
            </tr>
          </thead>
          <tbody>
            {invoices.map((invoice) => (
              <tr
                key={invoice.id}
                className="border-b border-stone-100 last:border-b-0"
              >
                <td className="px-4 py-3 align-top font-medium text-stone-900">
                  #{invoice.id}
                </td>
                <td className="break-words px-4 py-3 align-top text-stone-700">
                  <p>{invoiceCustomerLabel(invoice)}</p>
                  {invoice.createdByUser?.username ? (
                    <p className="mt-0.5 text-xs text-stone-400">
                      {invoice.createdByUser.username}
                    </p>
                  ) : null}
                </td>
                <td className="px-4 py-3 align-top text-right font-medium tabular-nums text-stone-900">
                  {formatMoney(invoice.totalAmount)}
                </td>
                <td className="px-4 py-3 align-top text-right tabular-nums text-stone-600">
                  {formatMoney(invoice.paidAmount)}
                </td>
                <td className="px-4 py-3 align-top text-right tabular-nums text-stone-600">
                  {formatMoney(invoice.remainingAmount)}
                </td>
                <td className="px-4 py-3 align-top">
                  <InvoiceStatusBadge status={invoice.status} />
                </td>
                <td className="px-4 py-3 align-top text-stone-500">
                  {invoice.createdAt}
                </td>
                <td className="px-4 py-3 align-top">
                  <div className="flex flex-wrap gap-1.5">
                    <Link
                      href={`/invoices/${invoice.id}`}
                      className="rounded-lg border border-stone-200 px-2.5 py-1.5 text-xs font-medium text-stone-700 hover:bg-stone-50"
                    >
                      Detalji
                    </Link>
                    <a
                      href={`/invoices/${invoice.id}/print?autoprint=1`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="rounded-lg border border-stone-200 px-2.5 py-1.5 text-xs font-medium text-stone-700 hover:bg-stone-50"
                    >
                      Štampaj
                    </a>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
