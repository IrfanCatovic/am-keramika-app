"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import {
  fetchCustomerPayments,
  getApiBusinessMessage,
} from "@/lib/customers-api";
import { formatMoney } from "@/lib/format";
import { CustomerPayment } from "@/types/customer";

export function CustomerPayments({
  customerId,
  limit = 5,
}: {
  customerId: number;
  limit?: number;
}) {
  const [payments, setPayments] = useState<CustomerPayment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      try {
        const data = await fetchCustomerPayments(customerId);
        if (cancelled) {
          return;
        }
        setPayments(data.slice(0, limit));
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
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [customerId, limit, reloadToken]);

  return (
    <section className="rounded-2xl border border-stone-200 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="border-b border-stone-100 px-4 py-3.5 sm:px-5">
        <h2 className="text-base font-semibold text-stone-900">
          Poslednje uplate
        </h2>
        <p className="mt-0.5 text-sm text-stone-500">
          Istorija evidentiranih uplata
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
        {!loading && !error && payments.length === 0 ? (
          <EmptyState
            title="Nema uplata"
            description="Za ovog kupca još nema evidentiranih uplata."
          />
        ) : null}
        {!loading && !error && payments.length > 0 ? (
          <ul className="space-y-3">
            {payments.map((payment) => (
              <li key={payment.id}>
                <Link
                  href={`/payments/${payment.id}`}
                  className="block rounded-xl border border-stone-200 px-3 py-3 transition hover:border-[#c4a484]/50 hover:bg-[#faf7f3]"
                >
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <p className="font-medium text-stone-900">
                        Uplata #{payment.id}
                      </p>
                      <p className="mt-1 text-xs text-stone-500">
                        {payment.createdAt}
                      </p>
                    </div>
                    <p className="text-sm font-semibold text-stone-900">
                      {formatMoney(payment.totalAmount)}
                    </p>
                  </div>
                  {payment.allocations?.length ? (
                    <ul className="mt-2 space-y-1 border-t border-stone-100 pt-2 text-xs text-stone-500">
                      {payment.allocations.map((allocation) => (
                        <li key={allocation.id}>
                          Račun #{allocation.invoiceID}:{" "}
                          {formatMoney(allocation.amount)}
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </Link>
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  );
}
