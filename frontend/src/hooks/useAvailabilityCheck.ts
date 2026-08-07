"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import {
  checkPublicProductAvailability,
  PublicAvailabilityError,
} from "@/lib/public-availability-api";

const INSUFFICIENT_MESSAGE = "Nema dovoljno proizvoda na stanju.";
const NETWORK_MESSAGE =
  "Trenutno nije moguće proveriti dostupnost. Pokušajte ponovo.";

export function useAvailabilityCheck() {
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const requestId = useRef(0);

  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  const clearError = useCallback(() => setError(null), []);

  const checkNow = useCallback(async (productId: number, quantity: number) => {
    const id = ++requestId.current;
    setChecking(true);
    setError(null);
    try {
      const result = await checkPublicProductAvailability(productId, quantity);
      if (id !== requestId.current) return { available: false, stale: true as const };
      if (!result.available) {
        setError(INSUFFICIENT_MESSAGE);
        return { available: false, stale: false as const };
      }
      return { available: true, stale: false as const };
    } catch (err) {
      if (id !== requestId.current) return { available: false, stale: true as const };
      const message =
        err instanceof PublicAvailabilityError
          ? err.message
          : NETWORK_MESSAGE;
      setError(message);
      return { available: false, stale: false as const };
    } finally {
      if (id === requestId.current) setChecking(false);
    }
  }, []);

  const checkDebounced = useCallback(
    (productId: number, quantity: number, delayMs = 350) => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
      debounceRef.current = setTimeout(() => {
        void checkNow(productId, quantity);
      }, delayMs);
    },
    [checkNow],
  );

  return { checking, error, setError, clearError, checkNow, checkDebounced };
}

export { INSUFFICIENT_MESSAGE };
