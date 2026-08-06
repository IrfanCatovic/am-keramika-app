"use client";

import { useCallback, useEffect, useState } from "react";

import { getErrorMessage } from "@/lib/dashboard";

export type SectionStatus = "loading" | "ready" | "error";

export function useAsyncSection<T>(
  loader: () => Promise<T>,
  fallbackError: string,
  deps: readonly unknown[] = [],
) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<SectionStatus>("loading");
  const [reloadToken, setReloadToken] = useState(0);

  const retry = useCallback(() => {
    setReloadToken((value) => value + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      setStatus("loading");
      setError(null);
      try {
        const result = await loader();
        if (cancelled) {
          return;
        }
        setData(result);
        setStatus("ready");
      } catch (err) {
        if (cancelled) {
          return;
        }
        setData(null);
        setError(getErrorMessage(err, fallbackError));
        setStatus("error");
      }
    }

    void run();

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- deps controlled by caller
  }, [reloadToken, fallbackError, ...deps]);

  return { data, error, status, retry };
}
