"use client";

import { useCallback, useRef, useState } from "react";

import { ApiError } from "@/lib/api";
import {
  downloadInvoicePdf,
  shareOrDownloadInvoicePdf,
} from "@/lib/invoice-pdf-api";

export function useInvoicePdf() {
  const [downloadLoadingId, setDownloadLoadingId] = useState<number | null>(
    null,
  );
  const [shareLoadingId, setShareLoadingId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);
  const busyRef = useRef(false);

  const clearMessages = useCallback(() => {
    setError(null);
    setInfo(null);
  }, []);

  const download = useCallback(async (invoiceId: number) => {
    if (busyRef.current) {
      return;
    }
    busyRef.current = true;
    setDownloadLoadingId(invoiceId);
    setError(null);
    setInfo(null);
    try {
      await downloadInvoicePdf(invoiceId);
    } catch (err) {
      const message =
        err instanceof ApiError
          ? err.message
          : "Preuzimanje PDF-a nije uspjelo.";
      setError(message);
    } finally {
      setDownloadLoadingId(null);
      busyRef.current = false;
    }
  }, []);

  const share = useCallback(async (invoiceId: number) => {
    if (busyRef.current) {
      return;
    }
    busyRef.current = true;
    setShareLoadingId(invoiceId);
    setError(null);
    setInfo(null);
    try {
      const result = await shareOrDownloadInvoicePdf(invoiceId);
      if (result.mode === "downloaded") {
        setInfo(result.message);
      }
    } catch (err) {
      const message =
        err instanceof ApiError
          ? err.message
          : "Dijeljenje PDF-a nije uspjelo.";
      setError(message);
    } finally {
      setShareLoadingId(null);
      busyRef.current = false;
    }
  }, []);

  return {
    download,
    share,
    downloadLoadingId,
    shareLoadingId,
    error,
    info,
    clearMessages,
    isBusy: downloadLoadingId != null || shareLoadingId != null,
  };
}
