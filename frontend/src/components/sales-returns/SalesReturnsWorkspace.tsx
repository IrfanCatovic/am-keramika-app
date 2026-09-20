"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import { formatMoney } from "@/lib/format";
import {
  fetchSalesReturns,
  getApiBusinessMessage,
} from "@/lib/sales-return-api";
import { userDisplayName } from "@/lib/user-display";
import {
  PaginatedSalesReturns,
  SalesReturnListItem,
} from "@/types/sales-return";

type CashFilter = "all" | "true" | "false";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function parseCashFilter(value: string | null): CashFilter {
  if (value === "true" || value === "false") {
    return value;
  }
  return "all";
}

function itemsCountLabel(count: number): string {
  if (count === 1) {
    return "1 artikal";
  }
  if (count >= 2 && count <= 4) {
    return `${count} artikla`;
  }
  return `${count} artikala`;
}

function CashBadge({ cashRefunded }: { cashRefunded: boolean }) {
  if (cashRefunded) {
    return (
      <span className="inline-flex rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-0.5 text-xs font-medium text-emerald-800">
        Novac vraćen
      </span>
    );
  }
  return (
    <span className="inline-flex rounded-full border border-stone-200 bg-stone-50 px-2.5 py-0.5 text-xs font-medium text-stone-600">
      Bez isplate
    </span>
  );
}

export function SalesReturnsWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const searchFromUrl = searchParams.get("search") ?? "";
  const cashFilter = parseCashFilter(searchParams.get("cashRefunded"));

  const [searchInput, setSearchInput] = useState(searchFromUrl);
  const [data, setData] = useState<PaginatedSalesReturns | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const syncQuery = useCallback(
    (patch: {
      page?: number;
      search?: string;
      cashRefunded?: CashFilter;
    }) => {
      const params = new URLSearchParams(searchParams.toString());

      const nextPage = patch.page ?? page;
      if (nextPage > 1) {
        params.set("page", String(nextPage));
      } else {
        params.delete("page");
      }

      const nextSearch =
        patch.search !== undefined ? patch.search : searchFromUrl;
      const trimmed = nextSearch.trim();
      if (trimmed) {
        params.set("search", trimmed);
      } else {
        params.delete("search");
      }

      const nextCash =
        patch.cashRefunded !== undefined ? patch.cashRefunded : cashFilter;
      if (nextCash === "true" || nextCash === "false") {
        params.set("cashRefunded", nextCash);
      } else {
        params.delete("cashRefunded");
      }

      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [cashFilter, page, pathname, router, searchFromUrl, searchParams],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (searchInput.trim() === searchFromUrl.trim()) {
        return;
      }
      syncQuery({ search: searchInput, page: 1 });
    }, 350);
    return () => window.clearTimeout(timer);
  }, [searchInput, searchFromUrl, syncQuery]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchSalesReturns({
        page,
        pageSize: 20,
        search: searchFromUrl || undefined,
        cashRefunded:
          cashFilter === "all" ? undefined : cashFilter === "true",
      });
      setData(result);
    } catch (err) {
      setData(null);
      setError(getApiBusinessMessage(err, "Povrati robe nisu učitani."));
    } finally {
      setLoading(false);
    }
  }, [cashFilter, page, searchFromUrl]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void load();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [load, reloadToken]);

  const items: SalesReturnListItem[] = data?.items ?? [];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
            Povrati robe
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Evidencija robe vraćene na lager.
          </p>
        </div>
        <Link
          href="/inventory"
          className="inline-flex min-h-11 shrink-0 items-center justify-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Otvori Lager
        </Link>
      </div>

      <div className="rounded-2xl border border-stone-200 bg-white p-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <div>
            <label
              htmlFor="sales-returns-search"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Pretraži povrate
            </label>
            <input
              id="sales-returns-search"
              type="search"
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder="Opis ili naziv proizvoda"
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </div>
          <div>
            <label
              htmlFor="sales-returns-cash"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Isplata novca
            </label>
            <select
              id="sales-returns-cash"
              value={cashFilter}
              onChange={(event) =>
                syncQuery({
                  cashRefunded: parseCashFilter(event.target.value),
                  page: 1,
                })
              }
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            >
              <option value="all">Sve</option>
              <option value="true">Novac vraćen</option>
              <option value="false">Bez isplate</option>
            </select>
          </div>
        </div>
      </div>

      {error ? (
        <InlineError
          message={error}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
      ) : loading ? (
        <ListSkeleton rows={5} />
      ) : items.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-5 py-10 text-center">
          <p className="text-base font-medium text-stone-800">
            Još nema evidentiranih povrata robe.
          </p>
          <p className="mt-2 text-sm text-stone-500">
            Povrat robe možete evidentirati iz modula Lager.
          </p>
          <Link
            href="/inventory"
            className="mt-5 inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800"
          >
            Otvori Lager
          </Link>
        </div>
      ) : (
        <>
          <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr className="border-b border-stone-200 bg-stone-50/80 text-left text-xs uppercase tracking-[0.08em] text-stone-500">
                  <th className="px-4 py-3 font-semibold">Datum</th>
                  <th className="px-4 py-3 font-semibold">Opis</th>
                  <th className="px-4 py-3 font-semibold">Artikala</th>
                  <th className="px-4 py-3 font-semibold">Vrednost povrata</th>
                  <th className="px-4 py-3 font-semibold">Novac vraćen</th>
                  <th className="px-4 py-3 font-semibold">Evidentirao</th>
                  <th className="px-4 py-3 font-semibold">Akcija</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr
                    key={item.id}
                    className="border-b border-stone-100 last:border-b-0"
                  >
                    <td className="px-4 py-3 whitespace-nowrap text-stone-600">
                      {item.createdAt}
                    </td>
                    <td className="max-w-xs px-4 py-3 text-stone-700">
                      <span className="line-clamp-2">
                        {item.description.trim() || "—"}
                      </span>
                    </td>
                    <td className="px-4 py-3 tabular-nums text-stone-700">
                      {item.itemsCount}
                    </td>
                    <td className="px-4 py-3 font-semibold tabular-nums text-stone-900">
                      {formatMoney(item.totalAmount)}
                    </td>
                    <td className="px-4 py-3">
                      <CashBadge cashRefunded={item.cashRefunded} />
                    </td>
                    <td className="px-4 py-3 text-stone-600">
                      {userDisplayName(item.createdByUser)}
                    </td>
                    <td className="px-4 py-3">
                      <Link
                        href={`/sales-returns/${item.id}`}
                        className="inline-flex min-h-9 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
                      >
                        Pogledaj
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <ul className="space-y-3 lg:hidden">
            {items.map((item) => (
              <li
                key={item.id}
                className="rounded-2xl border border-stone-200 bg-white p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-medium text-stone-900">
                      Povrat #{item.id}
                    </p>
                    <p className="mt-0.5 text-xs text-stone-500">
                      {item.createdAt}
                    </p>
                  </div>
                  <p className="shrink-0 font-semibold tabular-nums text-stone-900">
                    {formatMoney(item.totalAmount)}
                  </p>
                </div>
                <p className="mt-2 text-sm text-stone-600">
                  {itemsCountLabel(item.itemsCount)}
                </p>
                <div className="mt-2">
                  <CashBadge cashRefunded={item.cashRefunded} />
                </div>
                {item.description.trim() ? (
                  <p className="mt-2 line-clamp-2 text-sm text-stone-700">
                    {item.description}
                  </p>
                ) : null}
                <p className="mt-2 text-xs text-stone-500">
                  Evidentirao: {userDisplayName(item.createdByUser)}
                </p>
                <Link
                  href={`/sales-returns/${item.id}`}
                  className="mt-3 inline-flex min-h-10 w-full items-center justify-center rounded-xl border border-stone-200 text-sm font-medium text-stone-700 hover:bg-stone-50"
                >
                  Pogledaj
                </Link>
              </li>
            ))}
          </ul>
        </>
      )}

      {data && data.totalPages > 1 ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-stone-200 bg-white px-4 py-3 text-sm">
          <p className="text-stone-600">
            Stranica {data.page} / {data.totalPages}
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={data.page <= 1}
              onClick={() => syncQuery({ page: data.page - 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 font-medium text-stone-700 disabled:opacity-50"
            >
              Prethodna
            </button>
            <button
              type="button"
              disabled={data.page >= data.totalPages}
              onClick={() => syncQuery({ page: data.page + 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 font-medium text-stone-700 disabled:opacity-50"
            >
              Sledeća
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
