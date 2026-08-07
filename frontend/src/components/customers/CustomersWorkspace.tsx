"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { CustomerList } from "@/components/customers/CustomerList";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import {
  deleteCustomer,
  fetchCustomers,
  getApiBusinessMessage,
  updateCustomerStatus,
} from "@/lib/customers-api";
import { CustomerListItem } from "@/types/customer";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

type StatusFilter = "active" | "inactive" | "all";

function parseStatus(value: string | null): StatusFilter {
  if (value === "all" || value === "inactive") {
    return value;
  }
  return "active";
}

export function CustomersWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const limit = parsePositiveInt(searchParams.get("limit")) ?? 20;
  const status = parseStatus(searchParams.get("status"));
  const includeInactive = status === "all" || status === "inactive";
  const searchFromUrl = searchParams.get("search") ?? "";

  const [searchInput, setSearchInput] = useState(searchFromUrl);
  const [customers, setCustomers] = useState<CustomerListItem[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setSearchInput(searchFromUrl);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [searchFromUrl]);

  const [confirm, setConfirm] = useState<
    | { open: false }
    | {
        open: true;
        kind: "activate" | "deactivate" | "delete";
        customer: CustomerListItem;
      }
  >({ open: false });
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const syncQuery = useCallback(
    (patch: {
      page?: number;
      search?: string;
      status?: StatusFilter;
      limit?: number;
    }) => {
      const params = new URLSearchParams(searchParams.toString());
      const nextPage = patch.page ?? page;
      const nextSearch =
        patch.search !== undefined ? patch.search : searchFromUrl;
      const nextStatus = patch.status ?? status;
      const nextLimit = patch.limit ?? limit;

      if (nextPage > 1) {
        params.set("page", String(nextPage));
      } else {
        params.delete("page");
      }
      if (nextSearch.trim()) {
        params.set("search", nextSearch.trim());
      } else {
        params.delete("search");
      }
      if (nextStatus === "active") {
        params.delete("status");
      } else {
        params.set("status", nextStatus);
      }
      if (nextLimit !== 20) {
        params.set("limit", String(nextLimit));
      } else {
        params.delete("limit");
      }

      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [limit, page, pathname, router, searchFromUrl, searchParams, status],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (searchInput !== searchFromUrl) {
        syncQuery({ search: searchInput, page: 1 });
      }
    }, 350);
    return () => window.clearTimeout(timer);
  }, [searchInput, searchFromUrl, syncQuery]);

  const loadCustomers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetchCustomers({
        page,
        limit,
        search: searchFromUrl,
        includeInactive,
      });
      const rows = response.data ?? [];
      setCustomers(
        status === "inactive" ? rows.filter((item) => !item.isActive) : rows,
      );
      setTotal(response.total ?? 0);
      setTotalPages(Math.max(1, response.total_pages || 1));
    } catch (err) {
      setCustomers([]);
      setError(getApiBusinessMessage(err, "Nije moguće učitati kupce."));
    } finally {
      setLoading(false);
    }
  }, [includeInactive, limit, page, searchFromUrl, status]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadCustomers();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadCustomers, reloadToken]);

  async function handleConfirm() {
    if (!confirm.open) {
      return;
    }
    setConfirmLoading(true);
    setConfirmError(null);
    setBusyId(confirm.customer.id);
    try {
      if (confirm.kind === "delete") {
        await deleteCustomer(confirm.customer.id);
      } else {
        await updateCustomerStatus(
          confirm.customer.id,
          confirm.kind === "activate",
        );
      }
      setConfirm({ open: false });
      setReloadToken((value) => value + 1);
    } catch (err) {
      setConfirmError(
        getApiBusinessMessage(
          err,
          confirm.kind === "delete"
            ? "Kupac se ne može obrisati jer postoji istorija računa ili uplata. Možete ga deaktivirati ako nema otvorenih računa."
            : "Promjena statusa nije uspjela.",
        ),
      );
    } finally {
      setConfirmLoading(false);
      setBusyId(null);
    }
  }

  const confirmCopy = useMemo(() => {
    if (!confirm.open) {
      return {
        title: "",
        message: "",
        label: "Potvrdi",
        tone: "neutral" as const,
      };
    }
    if (confirm.kind === "delete") {
      return {
        title: "Obriši kupca",
        message: `Obrisati kupca „${confirm.customer.name}”? Brisanje nije moguće ako postoji istorija računa ili uplata.`,
        label: "Obriši",
        tone: "danger" as const,
      };
    }
    if (confirm.kind === "activate") {
      return {
        title: "Aktiviraj kupca",
        message: `Aktivirati kupca „${confirm.customer.name}”?`,
        label: "Aktiviraj",
        tone: "neutral" as const,
      };
    }
    return {
      title: "Deaktiviraj kupca",
      message: `Deaktivirati kupca „${confirm.customer.name}”? Nije dozvoljeno ako ima unpaid ili partially_paid račune.`,
      label: "Deaktiviraj",
      tone: "neutral" as const,
    };
  }, [confirm]);

  const inactiveOnly = status === "inactive";

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            AM Keramika
          </p>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
            Kupci
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Pregled i upravljanje kupcima.
          </p>
        </div>
        <Link
          href="/customers/new"
          className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800"
        >
          Novi kupac
        </Link>
      </header>

      <section className="dash-enter rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div className="sm:col-span-2">
            <label className="mb-1.5 block text-sm font-medium text-stone-700">
              Pretraga
            </label>
            <input
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder="Ime ili telefon"
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-stone-700">
              Status
            </label>
            <select
              value={status}
              onChange={(event) =>
                syncQuery({
                  status: event.target.value as StatusFilter,
                  page: 1,
                })
              }
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm"
            >
              <option value="active">Aktivni</option>
              <option value="inactive">Neaktivni</option>
              <option value="all">Svi</option>
            </select>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-stone-700">
              Po stranici
            </label>
            <select
              value={limit}
              onChange={(event) =>
                syncQuery({ limit: Number(event.target.value), page: 1 })
              }
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm"
            >
              <option value={10}>10</option>
              <option value={20}>20</option>
              <option value={50}>50</option>
            </select>
          </div>
        </div>
        {inactiveOnly ? (
          <p className="mt-3 text-xs text-stone-500">
            Prikazano {customers.length} neaktivnih na trenutnoj stranici.
            Filter „samo neaktivni“ nije server-side — ukupan broj neaktivnih
            nije dostupan.
          </p>
        ) : (
          <p className="mt-3 text-xs text-stone-500">Ukupno: {total}</p>
        )}
      </section>

      <CustomerList
        customers={customers}
        loading={loading}
        error={error}
        searchActive={Boolean(searchFromUrl.trim())}
        busyId={busyId}
        onRetry={() => setReloadToken((value) => value + 1)}
        onActivate={(customer) => {
          setConfirmError(null);
          setConfirm({ open: true, kind: "activate", customer });
        }}
        onDeactivate={(customer) => {
          setConfirmError(null);
          setConfirm({ open: true, kind: "deactivate", customer });
        }}
        onDelete={(customer) => {
          setConfirmError(null);
          setConfirm({ open: true, kind: "delete", customer });
        }}
      />

      {!inactiveOnly && totalPages > 1 ? (
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
              Sljedeća
            </button>
          </div>
        </div>
      ) : null}

      {inactiveOnly && totalPages > 1 ? (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-stone-500">
            Stranica {page} / {totalPages} (miješani aktivni/neaktivni
            rezultat)
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
              Sljedeća
            </button>
          </div>
        </div>
      ) : null}

      <ConfirmDialog
        open={confirm.open}
        title={confirmCopy.title}
        message={confirmCopy.message}
        confirmLabel={confirmCopy.label}
        tone={confirmCopy.tone}
        loading={confirmLoading}
        error={confirmError}
        onClose={() => {
          if (!confirmLoading) {
            setConfirm({ open: false });
          }
        }}
        onConfirm={() => void handleConfirm()}
      />
    </div>
  );
}
