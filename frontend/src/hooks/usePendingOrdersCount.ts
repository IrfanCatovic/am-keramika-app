"use client";

import { useCallback, useEffect, useState } from "react";

import {
  ONLINE_ORDERS_CHANGED_EVENT,
  fetchPendingCount,
} from "@/lib/online-orders-api";

const POLL_INTERVAL_MS = 45_000;

export function usePendingOrdersCount() {
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const data = await fetchPendingCount();
      setCount(typeof data.count === "number" ? data.count : 0);
    } catch {
      // Keep last known count on transient errors / unauthorized.
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    const frame = window.requestAnimationFrame(() => {
      if (!cancelled) void refresh();
    });

    const intervalId = window.setInterval(() => {
      void refresh();
    }, POLL_INTERVAL_MS);

    function onFocus() {
      void refresh();
    }

    function onChanged() {
      void refresh();
    }

    window.addEventListener("focus", onFocus);
    window.addEventListener(ONLINE_ORDERS_CHANGED_EVENT, onChanged);

    return () => {
      cancelled = true;
      window.cancelAnimationFrame(frame);
      window.clearInterval(intervalId);
      window.removeEventListener("focus", onFocus);
      window.removeEventListener(ONLINE_ORDERS_CHANGED_EVENT, onChanged);
    };
  }, [refresh]);

  return { count, refresh, loading };
}
