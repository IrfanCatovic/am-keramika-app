"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import {
  InvoiceFilters,
  InvoiceStatusFilter,
} from "@/components/invoices/InvoiceFilters";
import { InvoiceList } from "@/components/invoices/InvoiceList";
import {
  fetchInvoices,
  getApiBusinessMessage,
} from "@/lib/invoices-api";
import {
  InvoiceListItem,
  InvoiceSort,
  InvoiceSortDirection,
  InvoiceStatus,
} from "@/types/invoice";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function parseStatus(value: string | null): InvoiceStatusFilter {
  if (
    value === "paid" ||
    value === "unpaid" ||
    value === "partially_paid" ||
    value === "cancelled"
  ) {
    return value;
  }
  return "";
}

function parseSort(value: string | null): InvoiceSort {
  return value === "totalAmount" ? "totalAmount" : "createdAt";
}

function parseDirection(value: string | null): InvoiceSortDirection {
  return value === "asc" ? "asc" : "desc";
}

export function InvoicesWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const limit = parsePositiveInt(searchParams.get("limit")) ?? 20;
  const status = parseStatus(searchParams.get("status"));
  const fromDate = searchParams.get("fromDate") ?? "";
  const toDate = searchParams.get("toDate") ?? "";
  const searchFromUrl = searchParams.get("search") ?? "";
  const sort = parseSort(searchParams.get("sort"));
  const direction = parseDirection(searchParams.get("direction"));
  const customerID = parsePositiveInt(searchParams.get("customerID")) ?? undefined;

  const [searchInput, setSearchInput] = useState(searchFromUrl);
  const [invoices, setInvoices] = useState<InvoiceListItem[]>([]);
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
      status?: InvoiceStatusFilter;
      fromDate?: string;
      toDate?: string;
      sort?: InvoiceSort;
      direction?: InvoiceSortDirection;
      limit?: number;
      customerID?: number | null;
    }) => {
      const params = new URLSearchParams();
      const nextPage = patch.page ?? page;
      const nextSearch =
        patch.search !== undefined ? patch.search : searchFromUrl;
      const nextStatus = patch.status !== undefined ? patch.status : status;
      const nextFrom =
        patch.fromDate !== undefined ? patch.fromDate : fromDate;
      const nextTo = patch.toDate !== undefined ? patch.toDate : toDate;
      const nextSort = patch.sort ?? sort;
      const nextDirection = patch.direction ?? direction;
      const nextLimit = patch.limit ?? limit;
      const nextCustomerID =
        patch.customerID !== undefined ? patch.customerID : customerID;

      if (nextPage > 1) {
        params.set("page", String(nextPage));
      }
      if (nextSearch.trim()) {
        params.set("search", nextSearch.trim());
      }
      if (nextStatus) {
        params.set("status", nextStatus);
      }
      if (nextFrom.trim()) {
        params.set("fromDate", nextFrom.trim());
      }
      if (nextTo.trim()) {
        params.set("toDate", nextTo.trim());
      }
      if (nextSort !== "createdAt") {
        params.set("sort", nextSort);
      }
      if (nextDirection !== "desc") {
        params.set("direction", nextDirection);
      }
      if (nextLimit !== 20) {
        params.set("limit", String(nextLimit));
      }
      if (nextCustomerID && nextCustomerID > 0) {
        params.set("customerID", String(nextCustomerID));
      }

      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [
      customerID,
      direction,
      fromDate,
      limit,
      page,
      pathname,
      router,
      searchFromUrl,
      sort,
      status,
      toDate,
    ],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (searchInput !== searchFromUrl) {
        syncQuery({ search: searchInput, page: 1 });
      }
    }, 350);
    return () => window.clearTimeout(timer);
  }, [searchInput, searchFromUrl, syncQuery]);

  const loadInvoices = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetchInvoices({
        page,
        limit,
        status: (status || undefined) as InvoiceStatus | undefined,
        fromDate: fromDate || undefined,
        toDate: toDate || undefined,
        search: searchFromUrl || undefined,
        sort,
        direction,
        customerID,
      });
      setInvoices(response.data ?? []);
      setTotal(response.total ?? 0);
      setTotalPages(Math.max(1, response.totalPages || 1));
    } catch (err) {
      setInvoices([]);
      setError(getApiBusinessMessage(err, "Nije moguće učitati račune."));
    } finally {
      setLoading(false);
    }
  }, [
    customerID,
    direction,
    fromDate,
    limit,
    page,
    searchFromUrl,
    sort,
    status,
    toDate,
  ]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadInvoices();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadInvoices, reloadToken]);

  const filtersActive = Boolean(
    searchFromUrl.trim() ||
      status ||
      fromDate ||
      toDate ||
      customerID ||
      sort !== "createdAt" ||
      direction !== "desc",
  );

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            AM Keramika
          </p>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
            Računi
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Kreiranje, pregled i storniranje računa.
          </p>
        </div>
        <Link
          href="/invoices/new"
          className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800"
        >
          Novi račun
        </Link>
      </header>

      <InvoiceFilters
        search={searchInput}
        status={status}
        fromDate={fromDate}
        toDate={toDate}
        sort={sort}
        direction={direction}
        onSearchChange={setSearchInput}
        onStatusChange={(value) => syncQuery({ status: value, page: 1 })}
        onFromDateChange={(value) => syncQuery({ fromDate: value, page: 1 })}
        onToDateChange={(value) => syncQuery({ toDate: value, page: 1 })}
        onSortChange={(value) => syncQuery({ sort: value, page: 1 })}
        onDirectionChange={(value) => syncQuery({ direction: value, page: 1 })}
        onReset={() => {
          setSearchInput("");
          router.replace(pathname, { scroll: false });
        }}
      />

      <p className="text-xs text-stone-500">Ukupno: {total}</p>

      <InvoiceList
        invoices={invoices}
        loading={loading}
        error={error}
        filtersActive={filtersActive}
        onRetry={() => setReloadToken((value) => value + 1)}
      />

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
