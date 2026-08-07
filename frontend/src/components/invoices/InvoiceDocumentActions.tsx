"use client";

import Link from "next/link";

import { useInvoicePdf } from "@/hooks/useInvoicePdf";

type Variant = "inline" | "stack" | "menu";

export function InvoiceDocumentActions({
  invoiceId,
  variant = "inline",
  showOpen = false,
  showShare = true,
  openLabel = "Otvori",
  printLabel = "Štampaj",
  className = "",
}: {
  invoiceId: number;
  variant?: Variant;
  showOpen?: boolean;
  showShare?: boolean;
  openLabel?: string;
  printLabel?: string;
  className?: string;
}) {
  const {
    download,
    share,
    downloadLoadingId,
    shareLoadingId,
    error,
    info,
    clearMessages,
  } = useInvoicePdf();

  const downloading = downloadLoadingId === invoiceId;
  const sharing = shareLoadingId === invoiceId;
  const busy = downloading || sharing;

  const baseBtn =
    "inline-flex min-h-10 items-center justify-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-50";
  const primaryBtn =
    "inline-flex min-h-10 items-center justify-center rounded-xl bg-stone-900 px-3 text-sm font-semibold text-white hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-50";

  const layout =
    variant === "stack"
      ? "flex flex-col gap-2"
      : "flex flex-wrap gap-2";

  return (
    <div className={`min-w-0 ${className}`}>
      <div className={layout}>
        <button
          type="button"
          disabled={busy}
          onClick={() => {
            clearMessages();
            void download(invoiceId);
          }}
          className={primaryBtn}
        >
          {downloading ? "Preuzimanje…" : "Preuzmi PDF"}
        </button>
        {showShare ? (
          <button
            type="button"
            disabled={busy}
            onClick={() => {
              clearMessages();
              void share(invoiceId);
            }}
            className={baseBtn}
          >
            {sharing ? "Dijeljenje…" : "Podijeli"}
          </button>
        ) : null}
        <a
          href={`/invoices/${invoiceId}/print?autoprint=1`}
          target="_blank"
          rel="noopener noreferrer"
          className={baseBtn}
        >
          {printLabel}
        </a>
        {showOpen ? (
          <Link href={`/invoices/${invoiceId}`} className={baseBtn}>
            {openLabel}
          </Link>
        ) : null}
      </div>
      {error ? (
        <p className="mt-2 break-words text-xs text-red-700">{error}</p>
      ) : null}
      {info ? (
        <p className="mt-2 break-words text-xs text-stone-600">{info}</p>
      ) : null}
    </div>
  );
}
