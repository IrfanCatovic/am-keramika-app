"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import { formatMoney } from "@/lib/format";
import {
  fetchRefunds,
  getApiBusinessMessage,
} from "@/lib/refunds-api";
import { userDisplayName } from "@/lib/user-display";
import {
  PaginatedRefunds,
  RefundListItem,
} from "@/types/refund";

function parsePositiveInt(value: string | null): number | null {
  if (!value) return null;
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

export function RefundsWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const invoiceID = parsePositiveInt(searchParams.get("invoiceID"));
  const customerID = parsePositiveInt(searchParams.get("customerID"));
  const fromDate = searchParams.get("fromDate") ?? "";
  const toDate = searchParams.get("toDate") ?? "";

  const [fromInput, setFromInput] = useState(fromDate);
  const [toInput, setToInput] = useState(toDate);
  const [invoiceInput, setInvoiceInput] = useState(
    invoiceID ? String(invoiceID) : "",
  );

  const [data, setData] = useState<PaginatedRefunds | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const syncQuery = useCallback(
    (patch: {
      page?: number;
      invoiceID?: number | null;
      customerID?: number | null;
      fromDate?: string;
      toDate?: string;
    }) => {
      const params = new URLSearchParams(searchParams.toString());
      const nextPage = patch.page ?? page;
      if (nextPage > 1) params.set("page", String(nextPage));
      else params.delete("page");

      const nextInvoice =
        patch.invoiceID !== undefined ? patch.invoiceID : invoiceID;
      if (nextInvoice) params.set("invoiceID", String(nextInvoice));
      else params.delete("invoiceID");

      const nextCustomer =
        patch.customerID !== undefined ? patch.customerID : customerID;
      if (nextCustomer) params.set("customerID", String(nextCustomer));
      else params.delete("customerID");

      const nextFrom = patch.fromDate !== undefined ? patch.fromDate : fromDate;
      if (nextFrom) params.set("fromDate", nextFrom);
      else params.delete("fromDate");

      const nextTo = patch.toDate !== undefined ? patch.toDate : toDate;
      if (nextTo) params.set("toDate", nextTo);
      else params.delete("toDate");

      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    },
    [customerID, fromDate, invoiceID, page, pathname, router, searchParams, toDate],
  );

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchRefunds({
        page,
        limit: 20,
        invoiceID: invoiceID ?? undefined,
        customerID: customerID ?? undefined,
        fromDate: fromDate || undefined,
        toDate: toDate || undefined,
      });
      setData(result);
    } catch (err) {
      setData(null);
      setError(getApiBusinessMessage(err, "Povrati nisu učitani."));
    } finally {
      setLoading(false);
    }
  }, [customerID, fromDate, invoiceID, page, toDate]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void load();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [load, reloadToken]);

  const refunds: RefundListItem[] = data?.refunds ?? [];
  const pagination = data?.pagination;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
            Povrati
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Pregled evidentiranih povrata nakon storniranja računa.
          </p>
        </div>
        <Link
          href="/reports"
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Nazad na izvještaje
        </Link>
      </div>

      <div className="rounded-2xl border border-stone-200 bg-white p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-stone-700">
              Od datuma
            </label>
            <input
              type="date"
              value={fromInput}
              onChange={(e) => setFromInput(e.target.value)}
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-stone-700">
              Do datuma
            </label>
            <input
              type="date"
              value={toInput}
              onChange={(e) => setToInput(e.target.value)}
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-stone-700">
              Broj računa
            </label>
            <input
              inputMode="numeric"
              value={invoiceInput}
              onChange={(e) => setInvoiceInput(e.target.value)}
              placeholder="npr. 42"
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </div>
          <div className="flex items-end gap-2">
            <button
              type="button"
              onClick={() =>
                syncQuery({
                  page: 1,
                  fromDate: fromInput,
                  toDate: toInput,
                  invoiceID: parsePositiveInt(invoiceInput),
                })
              }
              className="inline-flex min-h-11 flex-1 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white"
            >
              Primijeni
            </button>
            <button
              type="button"
              onClick={() => {
                setFromInput("");
                setToInput("");
                setInvoiceInput("");
                syncQuery({
                  page: 1,
                  fromDate: "",
                  toDate: "",
                  invoiceID: null,
                  customerID: null,
                });
              }}
              className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700"
            >
              Reset
            </button>
          </div>
        </div>
      </div>

      {error ? (
        <InlineError
          message={error}
          onRetry={() => setReloadToken((v) => v + 1)}
        />
      ) : loading ? (
        <ListSkeleton rows={5} />
      ) : refunds.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-5 py-10 text-center text-sm text-stone-500">
          Nema evidentiranih povrata za odabrane filtere.
        </div>
      ) : (
        <>
          <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr className="border-b border-stone-200 bg-stone-50/80 text-left text-xs uppercase tracking-[0.08em] text-stone-500">
                  <th className="px-4 py-3 font-semibold">Datum</th>
                  <th className="px-4 py-3 font-semibold">Račun</th>
                  <th className="px-4 py-3 font-semibold">Kupac</th>
                  <th className="px-4 py-3 font-semibold">Iznos</th>
                  <th className="px-4 py-3 font-semibold">Korisnik</th>
                </tr>
              </thead>
              <tbody>
                {refunds.map((refund) => (
                  <tr
                    key={refund.id}
                    className="border-b border-stone-100 last:border-b-0"
                  >
                    <td className="px-4 py-3 text-stone-600">
                      {refund.createdAt}
                    </td>
                    <td className="px-4 py-3">
                      <Link
                        href={`/invoices/${refund.invoiceID}`}
                        className="font-medium text-[#8a6a45] hover:text-stone-900"
                      >
                        #{refund.invoiceID}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-stone-700">
                      {refund.customerID ? (
                        <Link
                          href={`/customers/${refund.customerID}`}
                          className="hover:text-[#8a6a45]"
                        >
                          {refund.customerName ?? `Kupac #${refund.customerID}`}
                        </Link>
                      ) : (
                        refund.customerName ?? "Gotovina"
                      )}
                    </td>
                    <td className="px-4 py-3 font-semibold tabular-nums text-red-700">
                      −{formatMoney(refund.amount)}
                    </td>
                    <td className="px-4 py-3 text-stone-600">
                      {userDisplayName(refund.createdByUser)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <ul className="space-y-3 lg:hidden">
            {refunds.map((refund) => (
              <li
                key={refund.id}
                className="rounded-2xl border border-stone-200 bg-white p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-xs text-stone-500">{refund.createdAt}</p>
                    <Link
                      href={`/invoices/${refund.invoiceID}`}
                      className="mt-1 inline-block font-medium text-[#8a6a45]"
                    >
                      Račun #{refund.invoiceID}
                    </Link>
                  </div>
                  <p className="font-semibold tabular-nums text-red-700">
                    −{formatMoney(refund.amount)}
                  </p>
                </div>
                <p className="mt-2 text-sm text-stone-700">
                  {refund.customerID ? (
                    <Link href={`/customers/${refund.customerID}`}>
                      {refund.customerName ?? `Kupac #${refund.customerID}`}
                    </Link>
                  ) : (
                    refund.customerName ?? "Gotovina"
                  )}
                </p>
                <p className="mt-1 text-xs text-stone-500">
                  {userDisplayName(refund.createdByUser)}
                </p>
              </li>
            ))}
          </ul>
        </>
      )}

      {pagination && pagination.totalPages > 1 ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-stone-200 bg-white px-4 py-3 text-sm">
          <p className="text-stone-600">
            Stranica {pagination.page} / {pagination.totalPages}
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={pagination.page <= 1}
              onClick={() => syncQuery({ page: pagination.page - 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 disabled:opacity-50"
            >
              Prethodna
            </button>
            <button
              type="button"
              disabled={pagination.page >= pagination.totalPages}
              onClick={() => syncQuery({ page: pagination.page + 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 disabled:opacity-50"
            >
              Sljedeća
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
