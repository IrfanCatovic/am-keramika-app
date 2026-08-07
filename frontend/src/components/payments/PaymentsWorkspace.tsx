"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { CustomerSelector } from "@/components/customers/CustomerSelector";
import { PaymentCard } from "@/components/payments/PaymentCard";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { fetchCustomer } from "@/lib/customers-api";
import {
  fetchPayments,
  getApiBusinessMessage,
} from "@/lib/payments-api";
import { CustomerListItem } from "@/types/customer";
import { Payment } from "@/types/payment";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

export function PaymentsWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const limit = parsePositiveInt(searchParams.get("limit")) ?? 20;
  const fromDate = searchParams.get("fromDate") ?? "";
  const toDate = searchParams.get("toDate") ?? "";
  const customerID = parsePositiveInt(searchParams.get("customerID"));

  const [customer, setCustomer] = useState<CustomerListItem | null>(null);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    if (!customerID) {
      const timer = window.setTimeout(() => setCustomer(null), 0);
      return () => window.clearTimeout(timer);
    }
    let cancelled = false;
    void (async () => {
      try {
        const data = await fetchCustomer(customerID);
        if (!cancelled) {
          setCustomer({
            id: data.id,
            name: data.name,
            phone: data.phone,
            isActive: data.isActive,
          });
        }
      } catch {
        if (!cancelled) {
          setCustomer(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [customerID]);

  const updateQuery = useCallback(
    (patch: Record<string, string | null>) => {
      const next = new URLSearchParams(searchParams.toString());
      for (const [key, value] of Object.entries(patch)) {
        if (value == null || value === "") {
          next.delete(key);
        } else {
          next.set(key, value);
        }
      }
      const query = next.toString();
      router.replace(query ? `${pathname}?${query}` : pathname);
    },
    [pathname, router, searchParams],
  );

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      setLoading(true);
      try {
        const response = await fetchPayments({
          page,
          limit,
          customerID: customerID ?? undefined,
          fromDate: fromDate || undefined,
          toDate: toDate || undefined,
        });
        if (cancelled) {
          return;
        }
        setPayments(response.data ?? []);
        setTotal(response.total ?? 0);
        setTotalPages(Math.max(1, response.totalPages ?? 1));
        setError(null);
      } catch (err) {
        if (cancelled) {
          return;
        }
        setPayments([]);
        setError(getApiBusinessMessage(err, "Nije moguće učitati uplate."));
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [page, limit, customerID, fromDate, toDate, reloadToken]);

  const newHref = useMemo(() => {
    if (customerID) {
      return `/payments/new?customerID=${customerID}`;
    }
    return "/payments/new";
  }, [customerID]);

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
            Uplate
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Evidentirajte uplatu za jedan ili više otvorenih računa kupca.
          </p>
        </div>
        <Link
          href={newHref}
          className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800"
        >
          Nova uplata
        </Link>
      </header>

      <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,1.4fr)_auto_auto]">
          <CustomerSelector
            value={customer}
            onChange={(next) => {
              setCustomer(next);
              updateQuery({
                customerID: next ? String(next.id) : null,
                page: "1",
              });
            }}
            label="Filtriraj po kupcu"
          />
          <label className="block text-sm">
            <span className="mb-1.5 block font-medium text-stone-700">Od</span>
            <input
              type="date"
              value={fromDate}
              onChange={(event) =>
                updateQuery({ fromDate: event.target.value || null, page: "1" })
              }
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1.5 block font-medium text-stone-700">Do</span>
            <input
              type="date"
              value={toDate}
              onChange={(event) =>
                updateQuery({ toDate: event.target.value || null, page: "1" })
              }
              className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
            />
          </label>
        </div>
      </section>

      {loading ? <ListSkeleton rows={4} /> : null}
      {!loading && error ? (
        <InlineError
          message={error}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
      ) : null}
      {!loading && !error && payments.length === 0 ? (
        <EmptyState
          title="Nema uplata"
          description="Još nema evidentiranih uplata za izabrane filtere."
        />
      ) : null}
      {!loading && !error && payments.length > 0 ? (
        <div className="space-y-3">
          <p className="text-sm text-stone-500">
            Prikazano {payments.length} od {total}
          </p>
          <ul className="space-y-3">
            {payments.map((payment) => (
              <li key={payment.id}>
                <PaymentCard payment={payment} />
              </li>
            ))}
          </ul>
          {totalPages > 1 ? (
            <div className="flex items-center justify-between gap-3 pt-2">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => updateQuery({ page: String(page - 1) })}
                className="rounded-xl border border-stone-200 bg-white px-3 py-2 text-sm disabled:opacity-40"
              >
                Prethodna
              </button>
              <p className="text-sm text-stone-500">
                Strana {page} / {totalPages}
              </p>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => updateQuery({ page: String(page + 1) })}
                className="rounded-xl border border-stone-200 bg-white px-3 py-2 text-sm disabled:opacity-40"
              >
                Sledeća
              </button>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
