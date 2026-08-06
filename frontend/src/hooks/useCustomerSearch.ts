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
          const response = await searchActiveCustomers(trimmed, 20);
          if (cancelled) {
            return;
          }
          setResults(response.data ?? []);
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
