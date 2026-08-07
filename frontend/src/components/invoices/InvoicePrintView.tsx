"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";

import { InvoicePrintDocument } from "@/components/invoices/InvoicePrintDocument";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import {
  fetchInvoice,
  getApiBusinessMessage,
} from "@/lib/invoices-api";
import { InvoiceDetails } from "@/types/invoice";

function waitForImages(root: HTMLElement): Promise<void> {
  const images = Array.from(root.querySelectorAll("img"));
  if (images.length === 0) {
    return Promise.resolve();
  }
  return Promise.all(
    images.map(
      (img) =>
        new Promise<void>((resolve) => {
          if (img.complete) {
            resolve();
            return;
          }
          const done = () => resolve();
          img.addEventListener("load", done, { once: true });
          img.addEventListener("error", done, { once: true });
        }),
    ),
  ).then(() => undefined);
}

export function InvoicePrintView({
  invoiceId,
  autoprint,
}: {
  invoiceId: number;
  autoprint: boolean;
}) {
  const [invoice, setInvoice] = useState<InvoiceDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [printedAt, setPrintedAt] = useState<string>("");
  const [reloadToken, setReloadToken] = useState(0);
  const sheetRef = useRef<HTMLDivElement | null>(null);
  const printTriggeredRef = useRef(false);

  const loadInvoice = useCallback(async () => {
    setLoading(true);
    setError(null);
    printTriggeredRef.current = false;
    try {
      const data = await fetchInvoice(invoiceId);
      setInvoice(data);
      setPrintedAt(
        new Intl.DateTimeFormat("sr-RS", {
          dateStyle: "short",
          timeStyle: "short",
        }).format(new Date()),
      );
    } catch (err) {
      setInvoice(null);
      setError(getApiBusinessMessage(err, "Račun nije pronađen."));
    } finally {
      setLoading(false);
    }
  }, [invoiceId]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadInvoice();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadInvoice, reloadToken]);

  // Hide browser print chrome (date + page title like "8/7/26, 10:15 AM AM Keramika").
  useEffect(() => {
    const previousTitle = document.title;
    const blankTitle = () => {
      document.title = " ";
    };
    const restoreTitle = () => {
      document.title = previousTitle;
    };

    blankTitle();
    window.addEventListener("beforeprint", blankTitle);
    window.addEventListener("afterprint", restoreTitle);
    return () => {
      window.removeEventListener("beforeprint", blankTitle);
      window.removeEventListener("afterprint", restoreTitle);
      restoreTitle();
    };
  }, []);

  useEffect(() => {
    if (!autoprint || loading || error || !invoice || printTriggeredRef.current) {
      return;
    }

    let cancelled = false;
    const timer = window.setTimeout(() => {
      void (async () => {
        if (sheetRef.current) {
          await waitForImages(sheetRef.current);
        }
        await new Promise<void>((resolve) => {
          window.requestAnimationFrame(() => resolve());
        });
        if (cancelled || printTriggeredRef.current) {
          return;
        }
        printTriggeredRef.current = true;
        window.print();
      })();
    }, 120);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [autoprint, loading, error, invoice]);

  if (loading) {
    return (
      <div className="mx-auto max-w-[210mm] space-y-4 p-4">
        <ListSkeleton rows={4} />
      </div>
    );
  }

  if (error || !invoice) {
    return (
      <div className="mx-auto max-w-[210mm] space-y-4 p-4">
        <InlineError
          message={error ?? "Račun nije pronađen."}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
        <div className="flex flex-wrap gap-2">
          <Link
            href={`/invoices/${invoiceId}`}
            className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700"
          >
            Nazad na račun
          </Link>
          <Link
            href="/invoices/new"
            className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
          >
            Nova prodaja
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="invoice-print-page min-h-screen bg-stone-200/80">
      <div className="invoice-print-toolbar no-print mx-auto flex max-w-[210mm] flex-wrap items-center gap-2 px-3 py-3 sm:px-0">
        <Link
          href={`/invoices/${invoice.id}`}
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-300 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Nazad na račun
        </Link>
        <Link
          href="/invoices/new"
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-300 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Nova prodaja
        </Link>
        <button
          type="button"
          onClick={() => window.print()}
          className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800"
        >
          Štampaj / sačuvaj PDF
        </button>
      </div>

      <div ref={sheetRef} className="px-3 pb-8 sm:px-0">
        <InvoicePrintDocument invoice={invoice} printedAt={printedAt} />
      </div>
    </div>
  );
}
