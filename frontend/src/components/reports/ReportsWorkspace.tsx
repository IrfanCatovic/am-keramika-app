"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import { formatCount, formatMoney } from "@/lib/format";
import {
  addDays,
  endOfMonth,
  fetchDailyReport,
  fetchSalesSummaryReport,
  fetchTransactionsReport,
  getApiBusinessMessage,
  startOfMonth,
  toISODate,
} from "@/lib/reports-api";
import {
  DailyReport,
  FinancialTransaction,
  ReportRangePreset,
  SalesSummaryReport,
} from "@/types/report";

function parsePreset(value: string | null): ReportRangePreset {
  if (
    value === "today" ||
    value === "yesterday" ||
    value === "this-month" ||
    value === "last-month" ||
    value === "custom"
  ) {
    return value;
  }
  return "today";
}

function resolveRange(
  preset: ReportRangePreset,
  customFrom: string,
  customTo: string,
): { fromDate: string; toDate: string } {
  const today = new Date();
  today.setHours(12, 0, 0, 0);

  switch (preset) {
    case "yesterday": {
      const yesterday = addDays(today, -1);
      const iso = toISODate(yesterday);
      return { fromDate: iso, toDate: iso };
    }
    case "this-month":
      return {
        fromDate: toISODate(startOfMonth(today)),
        toDate: toISODate(today),
      };
    case "last-month": {
      const lastMonthAnchor = new Date(
        today.getFullYear(),
        today.getMonth() - 1,
        15,
      );
      return {
        fromDate: toISODate(startOfMonth(lastMonthAnchor)),
        toDate: toISODate(endOfMonth(lastMonthAnchor)),
      };
    }
    case "custom":
      return {
        fromDate: customFrom || toISODate(today),
        toDate: customTo || customFrom || toISODate(today),
      };
    case "today":
    default: {
      const iso = toISODate(today);
      return { fromDate: iso, toDate: iso };
    }
  }
}

function KpiCard({
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
      className={`min-w-0 rounded-2xl border p-4 ${
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
      {hint ? <p className="mt-1 text-xs leading-relaxed text-stone-500">{hint}</p> : null}
    </div>
  );
}

function formatTransactionDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("sr-RS", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(date);
}

export function ReportsWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const range = parsePreset(searchParams.get("range"));
  const fromFromUrl = searchParams.get("fromDate") ?? "";
  const toFromUrl = searchParams.get("toDate") ?? "";

  const [customFrom, setCustomFrom] = useState(fromFromUrl);
  const [customTo, setCustomTo] = useState(toFromUrl);

  const resolved = useMemo(
    () => resolveRange(range, fromFromUrl || customFrom, toFromUrl || customTo),
    [range, fromFromUrl, toFromUrl, customFrom, customTo],
  );

  const [summary, setSummary] = useState<SalesSummaryReport | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(true);
  const [summaryError, setSummaryError] = useState<string | null>(null);

  const [daily, setDaily] = useState<DailyReport | null>(null);
  const [dailyLoading, setDailyLoading] = useState(false);
  const [dailyError, setDailyError] = useState<string | null>(null);

  const [transactions, setTransactions] = useState<FinancialTransaction[]>([]);
  const [txLoading, setTxLoading] = useState(true);
  const [txError, setTxError] = useState<string | null>(null);

  const [reloadToken, setReloadToken] = useState(0);

  const syncRange = useCallback(
    (preset: ReportRangePreset, from?: string, to?: string) => {
      const params = new URLSearchParams();
      if (preset !== "today") {
        params.set("range", preset);
      }
      if (preset === "custom") {
        if (from) params.set("fromDate", from);
        if (to) params.set("toDate", to);
      }
      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [pathname, router],
  );

  useEffect(() => {
    let cancelled = false;
    const timer = window.setTimeout(() => {
      void (async () => {
        setSummaryLoading(true);
        setSummaryError(null);
        setTxLoading(true);
        setTxError(null);

        const isSingleDay = resolved.fromDate === resolved.toDate;

        if (isSingleDay) {
          setDailyLoading(true);
          setDailyError(null);
        } else {
          setDaily(null);
        }

        try {
          const summaryResult = await fetchSalesSummaryReport(
            resolved.fromDate,
            resolved.toDate,
          );
          if (!cancelled) setSummary(summaryResult);
        } catch (err) {
          if (!cancelled) {
            setSummary(null);
            setSummaryError(
              getApiBusinessMessage(err, "Sažetak nije učitan."),
            );
          }
        } finally {
          if (!cancelled) setSummaryLoading(false);
        }

        try {
          const txResult = await fetchTransactionsReport(
            resolved.fromDate,
            resolved.toDate,
          );
          if (!cancelled) setTransactions(txResult.transactions ?? []);
        } catch (err) {
          if (!cancelled) {
            setTransactions([]);
            setTxError(
              getApiBusinessMessage(err, "Promet nije učitan."),
            );
          }
        } finally {
          if (!cancelled) setTxLoading(false);
        }

        if (isSingleDay) {
          try {
            const dailyResult = await fetchDailyReport(resolved.fromDate);
            if (!cancelled) setDaily(dailyResult);
          } catch (err) {
            if (!cancelled) {
              setDaily(null);
              setDailyError(
                getApiBusinessMessage(err, "Dnevni pregled nije učitan."),
              );
            }
          } finally {
            if (!cancelled) setDailyLoading(false);
          }
        }
      })();
    }, 0);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [range, resolved.fromDate, resolved.toDate, reloadToken]);

  const presets: { id: ReportRangePreset; label: string }[] = [
    { id: "today", label: "Danas" },
    { id: "yesterday", label: "Juče" },
    { id: "this-month", label: "Ovaj mesec" },
    { id: "last-month", label: "Prošli mesec" },
    { id: "custom", label: "Prilagođeni" },
  ];

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
            Izveštaji
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Finansijski pregled za {resolved.fromDate}
            {resolved.fromDate !== resolved.toDate
              ? ` — ${resolved.toDate}`
              : ""}
          </p>
        </div>
        <Link
          href="/refunds"
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Pregled povrata
        </Link>
      </div>

      <div className="rounded-2xl border border-stone-200 bg-white p-3 sm:p-4">
        <div className="flex flex-wrap gap-2">
          {presets.map((preset) => (
            <button
              key={preset.id}
              type="button"
              onClick={() => {
                if (preset.id === "custom") {
                  syncRange(
                    "custom",
                    customFrom || resolved.fromDate,
                    customTo || resolved.toDate,
                  );
                  return;
                }
                syncRange(preset.id);
              }}
              className={`rounded-xl px-3 py-2 text-sm font-medium transition ${
                range === preset.id
                  ? "bg-stone-900 text-white"
                  : "border border-stone-200 text-stone-700 hover:bg-stone-50"
              }`}
            >
              {preset.label}
            </button>
          ))}
        </div>

        {range === "custom" ? (
          <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-stone-700">
                Od
              </label>
              <input
                type="date"
                value={customFrom || resolved.fromDate}
                onChange={(e) => setCustomFrom(e.target.value)}
                className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
              />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-stone-700">
                Do
              </label>
              <input
                type="date"
                value={customTo || resolved.toDate}
                onChange={(e) => setCustomTo(e.target.value)}
                className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
              />
            </div>
            <div className="flex items-end">
              <button
                type="button"
                onClick={() =>
                  syncRange(
                    "custom",
                    customFrom || resolved.fromDate,
                    customTo || resolved.toDate,
                  )
                }
                className="inline-flex min-h-11 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white"
              >
                Primeni period
              </button>
            </div>
          </div>
        ) : null}
      </div>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold uppercase tracking-[0.12em] text-stone-500">
          Pregled
        </h2>
        {summaryError ? (
          <InlineError
            message={summaryError}
            onRetry={() => setReloadToken((v) => v + 1)}
          />
        ) : summaryLoading ? (
          <ListSkeleton rows={2} />
        ) : summary ? (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <KpiCard
              label="Ukupna prodaja"
              value={formatMoney(summary.totalSales)}
              hint="Zbir računa u periodu (bez storniranih)."
            />
            <KpiCard
              label="Naplaćeno"
              value={formatMoney(summary.totalCollected)}
              hint="Stvarno evidentirane uplate u izabranom periodu."
            />
            <KpiCard
              label="Potraživanja (period)"
              value={formatMoney(summary.outstandingAmount)}
              hint="Neplaćeni dio računa kreiranih u periodu."
            />
            <KpiCard
              label="Povrati"
              value={formatMoney(summary.totalRefunds)}
              hint="Evidentirani povrati u periodu."
            />
            <KpiCard
              label="Neto promet"
              value={formatMoney(summary.netCash)}
              hint="Naplaćeno umanjeno za evidentirane povrate."
              accent
            />
            <KpiCard
              label="Broj računa"
              value={formatCount(summary.invoicesCount)}
              hint="Broj aktivnih (nestorniranih) računa u periodu."
            />
          </div>
        ) : null}
      </section>

      {resolved.fromDate === resolved.toDate ? (
        <section className="space-y-3">
          <h2 className="text-sm font-semibold uppercase tracking-[0.12em] text-stone-500">
            Dnevni tok
          </h2>
          {dailyError ? (
            <InlineError
              message={dailyError}
              onRetry={() => setReloadToken((v) => v + 1)}
            />
          ) : dailyLoading ? (
            <ListSkeleton rows={1} />
          ) : daily ? (
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              <KpiCard
                label="Uplata danas"
                value={formatCount(daily.paymentsCount)}
              />
              <KpiCard
                label="Povrata danas"
                value={formatCount(daily.refundsCount)}
              />
              <KpiCard
                label="Uplaćeno"
                value={formatMoney(daily.totalPayments)}
              />
              <KpiCard
                label="Neto cash"
                value={formatMoney(daily.netCash)}
                accent
              />
            </div>
          ) : null}
        </section>
      ) : null}

      <section className="space-y-3">
        <h2 className="text-sm font-semibold uppercase tracking-[0.12em] text-stone-500">
          Promet
        </h2>
        {txError ? (
          <InlineError
            message={txError}
            onRetry={() => setReloadToken((v) => v + 1)}
          />
        ) : txLoading ? (
          <ListSkeleton rows={4} />
        ) : transactions.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-5 py-10 text-center text-sm text-stone-500">
            Nema uplata ni povrata u odabranom periodu.
          </div>
        ) : (
          <>
            <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
              <table className="w-full border-collapse text-sm">
                <thead>
                  <tr className="border-b border-stone-200 bg-stone-50/80 text-left text-xs uppercase tracking-[0.08em] text-stone-500">
                    <th className="px-4 py-3 font-semibold">Datum</th>
                    <th className="px-4 py-3 font-semibold">Tip</th>
                    <th className="px-4 py-3 font-semibold">Iznos</th>
                    <th className="px-4 py-3 font-semibold">Kupac</th>
                    <th className="px-4 py-3 font-semibold">Račun</th>
                    <th className="px-4 py-3 font-semibold">Opis</th>
                  </tr>
                </thead>
                <tbody>
                  {transactions.map((tx) => {
                    const isRefund = tx.type === "refund";
                    const invoiceId = tx.invoiceIDs?.[0];
                    return (
                      <tr
                        key={`${tx.type}-${tx.id}`}
                        className="border-b border-stone-100 last:border-b-0"
                      >
                        <td className="px-4 py-3 text-stone-600">
                          {formatTransactionDate(tx.date)}
                        </td>
                        <td className="px-4 py-3 text-stone-800">
                          {isRefund ? "Povrat" : "Uplata"}
                        </td>
                        <td
                          className={`px-4 py-3 font-semibold tabular-nums ${
                            isRefund ? "text-red-700" : "text-emerald-800"
                          }`}
                        >
                          {isRefund ? "−" : "+"}
                          {formatMoney(tx.amount)}
                        </td>
                        <td className="px-4 py-3 text-stone-700">
                          {tx.customerID ? (
                            <Link
                              href={`/customers/${tx.customerID}`}
                              className="hover:text-[#8a6a45]"
                            >
                              {tx.customerName ?? `Kupac #${tx.customerID}`}
                            </Link>
                          ) : (
                            tx.customerName ?? "Gotovina"
                          )}
                        </td>
                        <td className="px-4 py-3">
                          {invoiceId ? (
                            <Link
                              href={`/invoices/${invoiceId}`}
                              className="font-medium text-[#8a6a45] hover:text-stone-900"
                            >
                              #{invoiceId}
                            </Link>
                          ) : !isRefund && tx.id ? (
                            <Link
                              href={`/payments/${tx.id}`}
                              className="font-medium text-[#8a6a45] hover:text-stone-900"
                            >
                              Uplata #{tx.id}
                            </Link>
                          ) : (
                            "—"
                          )}
                        </td>
                        <td className="px-4 py-3 text-stone-500">
                          {tx.description || "—"}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            <ul className="space-y-3 lg:hidden">
              {transactions.map((tx) => {
                const isRefund = tx.type === "refund";
                const invoiceId = tx.invoiceIDs?.[0];
                return (
                  <li
                    key={`${tx.type}-${tx.id}`}
                    className="rounded-2xl border border-stone-200 bg-white p-4"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-xs text-stone-500">
                          {formatTransactionDate(tx.date)}
                        </p>
                        <p className="mt-1 text-sm font-medium text-stone-900">
                          {isRefund ? "Povrat" : "Uplata"}
                        </p>
                      </div>
                      <p
                        className={`font-semibold tabular-nums ${
                          isRefund ? "text-red-700" : "text-emerald-800"
                        }`}
                      >
                        {isRefund ? "−" : "+"}
                        {formatMoney(tx.amount)}
                      </p>
                    </div>
                    <p className="mt-2 text-sm text-stone-700">
                      {tx.customerID ? (
                        <Link href={`/customers/${tx.customerID}`}>
                          {tx.customerName ?? `Kupac #${tx.customerID}`}
                        </Link>
                      ) : (
                        tx.customerName ?? "Gotovina"
                      )}
                    </p>
                    <div className="mt-2 flex flex-wrap gap-3 text-sm">
                      {invoiceId ? (
                        <Link
                          href={`/invoices/${invoiceId}`}
                          className="text-[#8a6a45]"
                        >
                          Račun #{invoiceId}
                        </Link>
                      ) : null}
                      {!isRefund ? (
                        <Link
                          href={`/payments/${tx.id}`}
                          className="text-[#8a6a45]"
                        >
                          Uplata #{tx.id}
                        </Link>
                      ) : null}
                    </div>
                  </li>
                );
              })}
            </ul>
          </>
        )}
      </section>
    </div>
  );
}
