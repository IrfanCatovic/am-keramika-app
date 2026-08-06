"use client";

import {
  SectionBody,
  SectionCard,
  SkeletonBlock,
} from "@/components/dashboard/SectionCard";
import { useAsyncSection } from "@/hooks/useAsyncSection";
import { fetchTodaysInvoices } from "@/lib/dashboard";

export function WorkerTodayStat({ date }: { date: string }) {
  const { data, error, status, retry } = useAsyncSection(
    () => fetchTodaysInvoices(date),
    "Nije moguće učitati današnje račune.",
    [date],
  );

  return (
    <SectionCard
      title="Današnji rad"
      description={`Računi kreirani ${date}`}
    >
      <SectionBody
        status={status}
        error={error}
        onRetry={retry}
        loadingFallback={<SkeletonBlock className="h-24" />}
      >
        <div className="rounded-2xl border border-stone-200 bg-gradient-to-br from-stone-50 to-white px-5 py-6">
          <p className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
            Broj današnjih računa
          </p>
          <p className="mt-2 text-3xl font-semibold tracking-tight text-stone-950">
            {new Intl.NumberFormat("bs-BA").format(data?.total ?? 0)}
          </p>
        </div>
      </SectionBody>
    </SectionCard>
  );
}
