"use client";

import { useCallback, useRef, useState } from "react";

import { ApiError } from "@/lib/api";
import { downloadInvoicePdf } from "@/lib/invoice-pdf-api";

export function useInvoicePdf() {
  const [downloadLoadingId, setDownloadLoadingId] = useState<number | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const busyRef = useRef(false);

  const clearMessages = useCallback(() => {
    setError(null);
  }, []);

  const download = useCallback(async (invoiceId: number) => {
    if (busyRef.current) {
      return;
    }
    busyRef.current = true;
    setDownloadLoadingId(invoiceId);
    setError(null);
    try {
      await downloadInvoicePdf(invoiceId);
    } catch (err) {
      const message =
        err instanceof ApiError
          ? err.message
          : "Preuzimanje PDF-a nije uspelo.";
      setError(message);
    } finally {
      setDownloadLoadingId(null);
      busyRef.current = false;
    }
  }, []);

  return {
    download,
    downloadLoadingId,
    error,
    clearMessages,
    isBusy: downloadLoadingId != null,
  };
}
