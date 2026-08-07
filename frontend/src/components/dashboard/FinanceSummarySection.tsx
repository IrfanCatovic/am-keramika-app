"use client";

import Link from "next/link";

import {
  SectionBody,
  SectionCard,
  SkeletonBlock,
} from "@/components/dashboard/SectionCard";
import { useAsyncSection } from "@/hooks/useAsyncSection";
import { fetchSalesSummary } from "@/lib/dashboard";
import { formatCount, formatMoney } from "@/lib/format";
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
      className={`dash-enter min-w-0 rounded-2xl border p-4 transition hover:-translate-y-0.5 hover:shadow-[0_8px_24px_rgba(28,25,23,0.06)] ${
        accent
          ? "border-[#c4a484]/45 bg-[#faf6f1]"
          : "border-stone-200/90 bg-white"
      }`}
    >
      <p className="text-[11px] font-medium uppercase tracking-[0.14em] text-stone-500">
        {label}
      </p>
      <p className="mt-2 break-words text-xl font-semibold tracking-tight text-stone-950 sm:text-2xl">
        {value}
      </p>
      {hint ? <p className="mt-1 text-xs text-stone-500">{hint}</p> : null}
    </div>
  );
}

function MetricsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
      {Array.from({ length: 6 }).map((_, index) => (
        <SkeletonBlock key={index} className="h-28" />
      ))}
    </div>
  );
}

function MetricsGrid({ summary }: { summary: SalesSummaryReport }) {
  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
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
        value={formatCount(summary.invoicesCount)}
      />
      <MetricCard
        label="Potraživanja (period)"
        value={formatMoney(summary.outstandingAmount)}
        hint="Neplaćeni dio računa kreiranih danas"
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
          href="/reports?range=today"
          className="shrink-0 text-sm font-medium text-[#8a6a45] transition hover:text-stone-900"
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
