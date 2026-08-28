"use client";

import { useEffect, useState } from "react";

import {
  getApiBusinessMessage,
  searchActiveCustomers,
} from "@/lib/customers-api";
import { CustomerListItem } from "@/types/customer";

/**
 * Debounced pretraga aktivnih kupaca — za CustomerSelector i buduću invoice formu.
 */
export function useCustomerSearch(query: string, enabled = true) {
  const [results, setResults] = useState<CustomerListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const trimmed = query.trim();
    let cancelled = false;
    const timer = window.setTimeout(() => {
      void (async () => {
        try {
          const pageSize = 50;
          const firstPage = await searchActiveCustomers(trimmed, pageSize, 1);
          const customers = [...(firstPage.data ?? [])];
          const totalPages = Math.max(1, firstPage.total_pages ?? 1);

          for (let page = 2; page <= totalPages; page += 1) {
            if (cancelled) {
              return;
            }
            const response = await searchActiveCustomers(
              trimmed,
              pageSize,
              page,
            );
            customers.push(...(response.data ?? []));
          }

          if (cancelled) {
            return;
          }
          setResults(customers.filter((item) => item.isActive));
          setError(null);
        } catch (err) {
          if (cancelled) {
            return;
          }
          setResults([]);
          setError(
            getApiBusinessMessage(err, "Nije moguće pretražiti kupce."),
          );
        } finally {
          if (!cancelled) {
            setLoading(false);
          }
        }
      })();
    }, 350);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [query, enabled]);

  useEffect(() => {
    if (!enabled) {
      return;
    }
    const timer = window.setTimeout(() => {
      setLoading(true);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [query, enabled]);

  return { results, loading, error };
}
