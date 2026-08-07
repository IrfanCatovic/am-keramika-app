"use client";

import Link from "next/link";

import { InvoiceStatusBadge } from "@/components/dashboard/InvoiceStatusBadge";
import {
  SectionBody,
  SectionCard,
  SectionEmpty,
  SkeletonBlock,
} from "@/components/dashboard/SectionCard";
import { useAsyncSection } from "@/hooks/useAsyncSection";
import { fetchRecentInvoices, invoiceCustomerLabel } from "@/lib/dashboard";
import { formatMoney } from "@/lib/format";
import { InvoiceListItem } from "@/types/invoice";

function InvoiceMobileCard({ invoice }: { invoice: InvoiceListItem }) {
  return (
    <li className="rounded-2xl border border-stone-200 bg-stone-50/70 p-3.5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="font-semibold text-stone-900">#{invoice.id}</p>
          <p className="mt-1 break-words text-sm text-stone-600">
            {invoiceCustomerLabel(invoice)}
          </p>
        </div>
        <InvoiceStatusBadge status={invoice.status} />
      </div>
      <div className="mt-3 flex items-end justify-between gap-3 border-t border-stone-200/80 pt-3">
        <p className="text-xs text-stone-500">{invoice.createdAt}</p>
        <p className="text-sm font-semibold tabular-nums text-stone-900">
          {formatMoney(invoice.totalAmount)}
        </p>
      </div>
    </li>
  );
}

function InvoiceDesktopRow({ invoice }: { invoice: InvoiceListItem }) {
  return (
    <tr className="border-b border-stone-100 last:border-b-0">
      <td className="py-3 pr-3 font-medium text-stone-900">#{invoice.id}</td>
      <td className="max-w-[14rem] truncate py-3 pr-3 text-sm text-stone-600">
        {invoiceCustomerLabel(invoice)}
      </td>
      <td className="py-3 pr-3 text-sm font-semibold tabular-nums text-stone-900">
        {formatMoney(invoice.totalAmount)}
      </td>
      <td className="py-3 pr-3">
        <InvoiceStatusBadge status={invoice.status} />
      </td>
      <td className="py-3 text-sm text-stone-500">{invoice.createdAt}</td>
    </tr>
  );
}

function InvoicesSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 4 }).map((_, index) => (
        <SkeletonBlock key={index} className="h-24 md:h-14" />
      ))}
    </div>
  );
}

export function RecentInvoicesSection() {
  const { data, error, status, retry } = useAsyncSection(
    () => fetchRecentInvoices(),
    "Nije moguće učitati poslednje račune.",
  );

  const invoices = data?.data ?? [];

  return (
    <SectionCard
      title="Poslednji računi"
      description="Najnoviji unosi u sistemu"
      action={
        <Link
          href="/invoices"
          className="shrink-0 text-sm font-medium text-[#8a6a45] transition hover:text-stone-900"
        >
          Svi računi
        </Link>
      }
    >
      <SectionBody
        status={status}
        error={error}
        onRetry={retry}
        loadingFallback={<InvoicesSkeleton />}
      >
        {invoices.length === 0 ? (
          <SectionEmpty message="Još nema kreiranih računa." />
        ) : (
          <>
            <ul className="space-y-3 md:hidden">
              {invoices.map((invoice) => (
                <InvoiceMobileCard key={invoice.id} invoice={invoice} />
              ))}
            </ul>

            <div className="hidden overflow-x-auto md:block">
              <table className="w-full min-w-[32rem] text-left text-sm">
                <thead>
                  <tr className="border-b border-stone-200 text-xs uppercase tracking-[0.08em] text-stone-500">
                    <th className="pb-2 pr-3 font-medium">Račun</th>
                    <th className="pb-2 pr-3 font-medium">Kupac</th>
                    <th className="pb-2 pr-3 font-medium">Iznos</th>
                    <th className="pb-2 pr-3 font-medium">Status</th>
                    <th className="pb-2 font-medium">Datum</th>
                  </tr>
                </thead>
                <tbody>
                  {invoices.map((invoice) => (
                    <InvoiceDesktopRow key={invoice.id} invoice={invoice} />
                  ))}
                </tbody>
              </table>
            </div>
          </>
        )}
      </SectionBody>
    </SectionCard>
  );
}
