"use client";

import Link from "next/link";

import {
  SectionBody,
  SectionCard,
  SkeletonBlock,
} from "@/components/dashboard/SectionCard";
import { useAsyncSection } from "@/hooks/useAsyncSection";
import { fetchSalesSummary, formatMoney } from "@/lib/dashboard";
import { SalesSummaryReport } from "@/types/report";

function MetricCard({
  label,
  value,
  hint,
  accent = false,
}: {
  label: string;
  value: string;
  hint?: string;
  accent?: boolean;
}) {
  return (
    <div
      className={`dash-enter rounded-2xl border p-4 transition hover:-translate-y-0.5 hover:shadow-md ${
        accent
          ? "border-[#c4a484]/40 bg-gradient-to-br from-[#f7f1ea] to-white"
          : "border-stone-200/80 bg-white"
      }`}
    >
      <p className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
        {label}
      </p>
      <p className="mt-2 text-xl font-semibold tracking-tight text-stone-950 sm:text-2xl">
        {value}
      </p>
      {hint ? <p className="mt-1 text-xs text-stone-500">{hint}</p> : null}
    </div>
  );
}

function MetricsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: 6 }).map((_, index) => (
        <SkeletonBlock key={index} className="h-28" />
      ))}
    </div>
  );
}

function MetricsGrid({ summary }: { summary: SalesSummaryReport }) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
      <MetricCard label="Ukupna prodaja" value={formatMoney(summary.totalSales)} />
      <MetricCard
        label="Naplaćeno"
        value={formatMoney(summary.totalCollected)}
      />
      <MetricCard
        label="Povrati novca"
        value={formatMoney(summary.totalRefunds)}
      />
      <MetricCard
        label="Neto promet"
        value={formatMoney(summary.netCash)}
        accent
      />
      <MetricCard
        label="Broj računa"
        value={new Intl.NumberFormat("bs-BA").format(summary.invoicesCount)}
      />
      <MetricCard
        label="Preostali dug"
        value={formatMoney(summary.outstandingAmount)}
        hint="Outstanding na današnji period"
      />
    </div>
  );
}

export function FinanceSummarySection({ date }: { date: string }) {
  const { data, error, status, retry } = useAsyncSection(
    () => fetchSalesSummary(date),
    "Nije moguće učitati današnji finansijski pregled.",
    [date],
  );

  return (
    <SectionCard
      title="Današnji pregled"
      description={`Finansijski sažetak za ${date}`}
      action={
        <Link
          href="/reports"
          className="text-sm font-medium text-[#8a6a45] transition hover:text-stone-900"
        >
          Izvještaji
        </Link>
      }
    >
      <SectionBody
        status={status}
        error={error}
        onRetry={retry}
        loadingFallback={<MetricsSkeleton />}
      >
        {data ? <MetricsGrid summary={data} /> : null}
      </SectionBody>
    </SectionCard>
  );
}
