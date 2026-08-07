"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { formatMoney } from "@/lib/format";
import {
  fetchOnlineOrders,
  formatOrderDateTime,
  formatRelativeReceived,
  getApiBusinessMessage,
  onlineOrderCustomerName,
  onlineOrderStatusLabel,
} from "@/lib/online-orders-api";
import {
  OnlineOrderListItem,
  OnlineOrderStatus,
} from "@/types/online-order-staff";

type StatusTab = "pending" | "confirmed" | "all";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function parseStatusTab(value: string | null): StatusTab {
  if (value === "confirmed") {
    return "confirmed";
  }
  if (value === "all") {
    return "all";
  }
  return "pending";
}

function statusBadgeClass(status: string): string {
  if (status === "pending") {
    return "bg-amber-50 text-amber-900 ring-amber-200";
  }
  if (status === "confirmed") {
    return "bg-emerald-50 text-emerald-800 ring-emerald-200";
  }
  return "bg-stone-100 text-stone-700 ring-stone-200";
}

const TABS: { key: StatusTab; label: string }[] = [
  { key: "pending", label: "Na čekanju" },
  { key: "confirmed", label: "Potvrđene" },
  { key: "all", label: "Sve" },
];

export function OrdersWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const limit = parsePositiveInt(searchParams.get("limit")) ?? 20;
  const statusTab = parseStatusTab(searchParams.get("status"));
  const searchFromUrl = searchParams.get("search") ?? "";

  const [searchInput, setSearchInput] = useState(searchFromUrl);
  const [orders, setOrders] = useState<OnlineOrderListItem[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setSearchInput(searchFromUrl);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [searchFromUrl]);

  const syncQuery = useCallback(
    (patch: {
      page?: number;
      search?: string;
      status?: StatusTab;
      limit?: number;
    }) => {
      const params = new URLSearchParams();
      const nextPage = patch.page ?? page;
      const nextSearch =
        patch.search !== undefined ? patch.search : searchFromUrl;
      const nextStatus = patch.status ?? statusTab;
      const nextLimit = patch.limit ?? limit;

      if (nextPage > 1) {
        params.set("page", String(nextPage));
      }
      if (nextSearch.trim()) {
        params.set("search", nextSearch.trim());
      }
      if (nextStatus !== "pending") {
        params.set("status", nextStatus);
      }
      if (nextLimit !== 20) {
        params.set("limit", String(nextLimit));
      }

      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [limit, page, pathname, router, searchFromUrl, statusTab],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (searchInput !== searchFromUrl) {
        syncQuery({ search: searchInput, page: 1 });
      }
    }, 350);
    return () => window.clearTimeout(timer);
  }, [searchInput, searchFromUrl, syncQuery]);

  const loadOrders = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const apiStatus: OnlineOrderStatus | "" =
        statusTab === "all" ? "" : statusTab;
      const response = await fetchOnlineOrders({
        page,
        limit,
        status: apiStatus,
        search: searchFromUrl || undefined,
      });
      setOrders(response.orders ?? []);
      setTotal(response.pagination?.totalItems ?? 0);
      setTotalPages(Math.max(1, response.pagination?.totalPages || 1));
    } catch (err) {
      setOrders([]);
      setError(getApiBusinessMessage(err, "Nije moguće učitati narudžbine."));
    } finally {
      setLoading(false);
    }
  }, [limit, page, searchFromUrl, statusTab]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadOrders();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadOrders, reloadToken]);

  const filtersActive = Boolean(
    searchFromUrl.trim() || statusTab !== "pending",
  );

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            AM Keramika
          </p>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
            Narudžbine
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Online narudžbine sa sajta — potvrda i kreiranje računa.
          </p>
        </div>
      </header>

      <section className="dash-enter rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div
            className="flex flex-wrap gap-1 rounded-xl bg-stone-100 p-1"
            role="tablist"
            aria-label="Status narudžbina"
          >
            {TABS.map((tab) => {
              const active = statusTab === tab.key;
              return (
                <button
                  key={tab.key}
                  type="button"
                  role="tab"
                  aria-selected={active}
                  onClick={() => syncQuery({ status: tab.key, page: 1 })}
                  className={`rounded-lg px-3 py-2 text-sm font-medium transition ${
                    active
                      ? "bg-white text-stone-900 shadow-sm"
                      : "text-stone-600 hover:text-stone-900"
                  }`}
                >
                  {tab.label}
                </button>
              );
            })}
          </div>
          <label className="block min-w-0 flex-1 sm:max-w-xs">
            <span className="mb-1.5 block text-sm font-medium text-stone-700">
              Pretraga
            </span>
            <input
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder="Ime, telefon ili grad"
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </label>
        </div>
      </section>

      <p className="text-xs text-stone-500">Ukupno: {total}</p>

      {loading ? <ListSkeleton rows={5} /> : null}
      {!loading && error ? (
        <InlineError
          message={error}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
      ) : null}
      {!loading && !error && orders.length === 0 ? (
        <EmptyState
          title={filtersActive ? "Nema rezultata" : "Nema narudžbina"}
          description={
            filtersActive
              ? "Pokušajte drugačije filtere ili pretragu."
              : "Kada stigne online narudžbina, pojaviće se ovde."
          }
        />
      ) : null}

      {!loading && !error && orders.length > 0 ? (
        <>
          <ul className="space-y-3 lg:hidden">
            {orders.map((order) => (
              <li key={order.id}>
                <article className="dash-enter min-w-0 rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div className="min-w-0">
                      <Link
                        href={`/orders/${order.id}`}
                        className="text-base font-semibold text-stone-900 hover:text-[#8a6a45]"
                      >
                        #{order.id}
                      </Link>
                      <p className="mt-1 break-words text-sm text-stone-600">
                        {onlineOrderCustomerName(order)}
                      </p>
                      <p className="mt-0.5 text-xs text-stone-500">
                        {order.phone}
                        {order.city ? ` · ${order.city}` : ""}
                      </p>
                    </div>
                    <span
                      className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${statusBadgeClass(order.status)}`}
                    >
                      {onlineOrderStatusLabel(order.status)}
                    </span>
                  </div>
                  <div className="mt-3 space-y-1 text-sm">
                    <p className="font-semibold tabular-nums text-stone-900">
                      {formatMoney(order.totalAmount)}
                    </p>
                    <p className="text-xs text-stone-500">
                      {formatRelativeReceived(order.createdAt)}
                      <span className="text-stone-400">
                        {" "}
                        · {formatOrderDateTime(order.createdAt)}
                      </span>
                    </p>
                  </div>
                  <div className="mt-4">
                    <Link
                      href={`/orders/${order.id}`}
                      className="inline-flex min-h-10 items-center justify-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
                    >
                      Detalj
                    </Link>
                  </div>
                </article>
              </li>
            ))}
          </ul>

          <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
            <table className="w-full table-fixed text-left text-sm">
              <thead className="sticky top-0 bg-stone-50/95 backdrop-blur">
                <tr className="border-b border-stone-200 text-xs uppercase tracking-[0.08em] text-stone-500">
                  <th className="w-[8%] px-3 py-3 font-medium">#</th>
                  <th className="w-[16%] px-3 py-3 font-medium">Kupac</th>
                  <th className="w-[12%] px-3 py-3 font-medium">Telefon</th>
                  <th className="w-[12%] px-3 py-3 font-medium">Grad</th>
                  <th className="w-[12%] px-3 py-3 font-medium text-right">
                    Vrednost
                  </th>
                  <th className="w-[12%] px-3 py-3 font-medium">Status</th>
                  <th className="w-[16%] px-3 py-3 font-medium">Primljena</th>
                  <th className="px-3 py-3 font-medium">Akcije</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr
                    key={order.id}
                    className="border-b border-stone-100 last:border-b-0"
                  >
                    <td className="px-3 py-3 align-top font-medium text-stone-900">
                      #{order.id}
                    </td>
                    <td className="break-words px-3 py-3 align-top text-stone-700">
                      {onlineOrderCustomerName(order)}
                    </td>
                    <td className="break-words px-3 py-3 align-top text-stone-600">
                      {order.phone}
                    </td>
                    <td className="break-words px-3 py-3 align-top text-stone-600">
                      {order.city}
                    </td>
                    <td className="px-3 py-3 align-top text-right font-medium tabular-nums text-stone-900">
                      {formatMoney(order.totalAmount)}
                    </td>
                    <td className="px-3 py-3 align-top">
                      <span
                        className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${statusBadgeClass(order.status)}`}
                      >
                        {onlineOrderStatusLabel(order.status)}
                      </span>
                    </td>
                    <td className="px-3 py-3 align-top text-stone-500">
                      <p className="text-stone-700">
                        {formatRelativeReceived(order.createdAt)}
                      </p>
                      <p className="mt-0.5 text-xs text-stone-400">
                        {formatOrderDateTime(order.createdAt)}
                      </p>
                    </td>
                    <td className="px-3 py-3 align-top">
                      <Link
                        href={`/orders/${order.id}`}
                        className="inline-flex min-h-10 items-center justify-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
                      >
                        Detalj
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : null}

      {totalPages > 1 ? (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-stone-500">
            Stranica {page} / {totalPages}
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => syncQuery({ page: page - 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm disabled:opacity-50"
            >
              Prethodna
            </button>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => syncQuery({ page: page + 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm disabled:opacity-50"
            >
              Sledeća
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
