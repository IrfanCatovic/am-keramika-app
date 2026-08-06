"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { InvoiceStatusBadge } from "@/components/dashboard/InvoiceStatusBadge";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import {
  fetchCustomerOpenInvoices,
  getApiBusinessMessage,
} from "@/lib/customers-api";
import { formatMoney } from "@/lib/format";
import { CustomerOpenInvoice } from "@/types/customer";

export function CustomerOpenInvoices({ customerId }: { customerId: number }) {
  const [invoices, setInvoices] = useState<CustomerOpenInvoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      try {
        const data = await fetchCustomerOpenInvoices(customerId);
        if (cancelled) {
          return;
        }
        setInvoices(data);
        setError(null);
      } catch (err) {
        if (cancelled) {
          return;
        }
        setInvoices([]);
        setError(
          getApiBusinessMessage(err, "Nije moguće učitati otvorene račune."),
        );
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [customerId, reloadToken]);

  return (
    <section className="rounded-2xl border border-stone-200 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="border-b border-stone-100 px-4 py-3.5 sm:px-5">
        <h2 className="text-base font-semibold text-stone-900">
          Otvoreni računi
        </h2>
        <p className="mt-0.5 text-sm text-stone-500">
          Neplaćeni i djelimično plaćeni računi
        </p>
      </div>
      <div className="px-4 py-4 sm:px-5">
        {loading ? <ListSkeleton rows={3} /> : null}
        {!loading && error ? (
          <InlineError
            message={error}
            onRetry={() => {
              setLoading(true);
              setReloadToken((value) => value + 1);
            }}
          />
        ) : null}
        {!loading && !error && invoices.length === 0 ? (
          <EmptyState
            title="Nema otvorenih računa"
            description="Kupac trenutno nema neizmirenih obaveza."
          />
        ) : null}
        {!loading && !error && invoices.length > 0 ? (
          <ul className="space-y-3">
            {invoices.map((invoice) => (
              <li
                key={invoice.id}
                className="rounded-xl border border-stone-200 px-3 py-3"
              >
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div className="min-w-0">
                    <Link
                      href={`/invoices/${invoice.id}`}
                      className="font-medium text-stone-900 hover:text-[#8a6a45]"
                    >
                      Račun #{invoice.id}
                    </Link>
                    <p className="mt-1 text-xs text-stone-500">
                      {invoice.createdAt}
                    </p>
                  </div>
                  <InvoiceStatusBadge status={invoice.status} />
                </div>
                <div className="mt-3 grid grid-cols-1 gap-1 text-sm text-stone-600 sm:grid-cols-3">
                  <p>
                    Ukupno:{" "}
                    <span className="font-medium text-stone-900">
                      {formatMoney(invoice.totalAmount)}
                    </span>
                  </p>
                  <p>
                    Plaćeno:{" "}
                    <span className="font-medium text-stone-900">
                      {formatMoney(invoice.paidAmount)}
                    </span>
                  </p>
                  <p>
                    Preostalo:{" "}
                    <span className="font-medium text-amber-900">
                      {formatMoney(invoice.remainingAmount)}
                    </span>
                  </p>
                </div>
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  );
}
