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
import {
  fetchRecentInvoices,
  formatMoney,
  invoiceCustomerLabel,
} from "@/lib/dashboard";
import { InvoiceListItem } from "@/types/invoice";

function InvoiceRow({ invoice }: { invoice: InvoiceListItem }) {
  return (
    <li className="flex flex-col gap-2 rounded-xl px-2 py-3 transition hover:bg-stone-50/80 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <p className="font-medium text-stone-900">#{invoice.id}</p>
          <InvoiceStatusBadge status={invoice.status} />
        </div>
        <p className="mt-1 truncate text-sm text-stone-600">
          {invoiceCustomerLabel(invoice)}
        </p>
        <p className="mt-0.5 text-xs text-stone-500">{invoice.createdAt}</p>
      </div>
      <p className="shrink-0 text-sm font-semibold text-stone-900 sm:text-right">
        {formatMoney(invoice.totalAmount)}
      </p>
    </li>
  );
}

function InvoicesSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 4 }).map((_, index) => (
        <SkeletonBlock key={index} className="h-16" />
      ))}
    </div>
  );
}

export function RecentInvoicesSection() {
  const { data, error, status, retry } = useAsyncSection(
    () => fetchRecentInvoices(),
    "Nije moguće učitati posljednje račune.",
  );

  const invoices = data?.data ?? [];

  return (
    <SectionCard
      title="Posljednji računi"
      description="Najnoviji unosi u sistemu"
      action={
        <Link
          href="/invoices"
          className="text-sm font-medium text-[#8a6a45] transition hover:text-stone-900"
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
          <ul className="divide-y divide-stone-100">
            {invoices.map((invoice) => (
              <InvoiceRow key={invoice.id} invoice={invoice} />
            ))}
          </ul>
        )}
      </SectionBody>
    </SectionCard>
  );
}
