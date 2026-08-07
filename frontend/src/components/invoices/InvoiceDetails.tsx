'use client';

import Link from 'next/link';
import { useCallback, useEffect, useState } from 'react';

import { CancelInvoiceDialog } from '@/components/invoices/CancelInvoiceDialog';
import { InvoiceDocumentActions } from '@/components/invoices/InvoiceDocumentActions';
import { InvoiceStatusBadge } from '@/components/invoices/InvoiceStatusBadge';
import { InlineError, ListSkeleton } from '@/components/ui/EmptyState';
import { formatMoney } from '@/lib/format';
import {
  cancelInvoice,
  fetchInvoice,
  getApiBusinessMessage,
  invoiceCustomerLabel,
} from '@/lib/invoices-api';
import { CancelInvoiceResponse, InvoiceDetails } from '@/types/invoice';

export function InvoiceDetailsView({ invoiceId }: { invoiceId: number }) {
  const [invoice, setInvoice] = useState<InvoiceDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const [cancelOpen, setCancelOpen] = useState(false);
  const [cancelLoading, setCancelLoading] = useState(false);
  const [cancelError, setCancelError] = useState<string | null>(null);
  const [lastCancel, setLastCancel] = useState<CancelInvoiceResponse | null>(
    null
  );

  const loadInvoice = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchInvoice(invoiceId);
      setInvoice(data);
    } catch (err) {
      setInvoice(null);
      setError(getApiBusinessMessage(err, 'Račun nije pronađen.'));
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

  async function handleCancel(reason: string) {
    if (!invoice || cancelLoading) {
      return;
    }
    setCancelLoading(true);
    setCancelError(null);
    try {
      const result = await cancelInvoice(invoice.id, { reason });
      setLastCancel(result);
      setCancelOpen(false);
      setReloadToken((value) => value + 1);
    } catch (err) {
      setCancelError(
        getApiBusinessMessage(err, 'Storniranje računa nije uspjelo.')
      );
    } finally {
      setCancelLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <ListSkeleton rows={2} />
        <ListSkeleton rows={4} />
      </div>
    );
  }

  if (error || !invoice) {
    return (
      <div className="space-y-4">
        <InlineError
          message={error ?? 'Račun nije pronađen.'}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
        <Link href="/invoices" className="text-sm font-medium text-[#8a6a45]">
          Nazad na račune
        </Link>
      </div>
    );
  }

  const isCancelled = invoice.status === 'cancelled';
  const isCash = invoice.customerID == null && !invoice.customer;
  const canRecordPayment =
    !isCancelled &&
    !isCash &&
    invoice.remainingAmount > 0 &&
    invoice.status !== 'paid';

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            Račun
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
              #{invoice.id}
            </h1>
            <InvoiceStatusBadge status={invoice.status} />
          </div>
          <p className="mt-1 text-sm text-stone-500">{invoice.createdAt}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link
            href="/invoices"
            className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
          >
            Nazad na račune
          </Link>
          <Link
            href="/invoices/new"
            className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
          >
            Novi račun
          </Link>
          {canRecordPayment ? (
            <Link
              href={`/payments/new?invoiceID=${invoice.id}`}
              className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
            >
              Evidentiraj uplatu
            </Link>
          ) : null}
          {!isCancelled ? (
            <button
              type="button"
              onClick={() => {
                setCancelError(null);
                setCancelOpen(true);
              }}
              className="inline-flex min-h-11 items-center rounded-xl border border-red-200 px-4 text-sm font-medium text-red-700 hover:bg-red-50"
            >
              Storniraj račun
            </button>
          ) : null}
        </div>
      </header>

      <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <h2 className="mb-3 text-sm font-semibold text-stone-900">Dokument</h2>
        <InvoiceDocumentActions
          invoiceId={invoice.id}
          printLabel="Štampaj račun"
        />
      </section>

      <section className="dash-enter rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
              Kupac
            </p>
            <p className="mt-1 break-words text-base font-semibold text-stone-900">
              {invoiceCustomerLabel(invoice)}
            </p>
            {invoice.customer?.phone ? (
              <p className="mt-1 text-sm text-stone-500">
                {invoice.customer.phone}
              </p>
            ) : null}
          </div>
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
              Kreirao
            </p>
            <p className="mt-1 text-base font-semibold text-stone-900">
              {invoice.createdByUser?.username ?? '—'}
            </p>
          </div>
        </div>

        <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div className="rounded-xl border border-stone-100 bg-stone-50 px-3 py-3">
            <p className="text-xs text-stone-500">Ukupno</p>
            <p className="mt-1 text-lg font-semibold tabular-nums text-stone-900">
              {formatMoney(invoice.totalAmount)}
            </p>
          </div>
          <div className="rounded-xl border border-stone-100 bg-stone-50 px-3 py-3">
            <p className="text-xs text-stone-500">Plaćeno</p>
            <p className="mt-1 text-lg font-semibold tabular-nums text-stone-900">
              {formatMoney(invoice.paidAmount)}
            </p>
          </div>
          <div className="rounded-xl border border-stone-100 bg-stone-50 px-3 py-3">
            <p className="text-xs text-stone-500">Preostalo</p>
            <p className="mt-1 text-lg font-semibold tabular-nums text-stone-900">
              {formatMoney(invoice.remainingAmount)}
            </p>
          </div>
        </div>
      </section>

      {isCancelled ? (
        <section className="rounded-2xl border border-amber-200 bg-amber-50/70 p-4 sm:p-5">
          <h2 className="text-sm font-semibold text-amber-950">
            Račun je storniran
          </h2>
          <div className="mt-3 grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
            <div>
              <p className="text-xs text-amber-800/80">Datum računa</p>
              <p className="mt-0.5 font-medium text-amber-950">
                {invoice.createdAt}
              </p>
            </div>
            {invoice.cancellation?.createdAt ? (
              <div>
                <p className="text-xs text-amber-800/80">Datum storna</p>
                <p className="mt-0.5 font-medium text-amber-950">
                  {invoice.cancellation.createdAt}
                </p>
              </div>
            ) : null}
            <div>
              <p className="text-xs text-amber-800/80">Originalni iznos</p>
              <p className="mt-0.5 font-semibold tabular-nums text-amber-950">
                {formatMoney(invoice.totalAmount)}
              </p>
            </div>
            <div>
              <p className="text-xs text-amber-800/80">Plaćeno prije storna</p>
              <p className="mt-0.5 font-semibold tabular-nums text-amber-950">
                {formatMoney(invoice.paidAmount)}
              </p>
            </div>
            {(invoice.cancellation?.createdByUser?.username ||
              invoice.refund?.createdByUser?.username) && (
              <div>
                <p className="text-xs text-amber-800/80">Stornirao</p>
                <p className="mt-0.5 font-medium text-amber-950">
                  {invoice.cancellation?.createdByUser?.username ??
                    invoice.refund?.createdByUser?.username}
                </p>
              </div>
            )}
            {invoice.cancellation?.reason ? (
              <div className="sm:col-span-2">
                <p className="text-xs text-amber-800/80">Razlog</p>
                <p className="mt-0.5 text-amber-950">
                  {invoice.cancellation.reason}
                </p>
              </div>
            ) : lastCancel?.reason ? (
              <div className="sm:col-span-2">
                <p className="text-xs text-amber-800/80">Razlog</p>
                <p className="mt-0.5 text-amber-950">{lastCancel.reason}</p>
              </div>
            ) : null}
          </div>
          {invoice.refund ? (
            <p className="mt-4 rounded-xl border border-amber-200 bg-white/70 px-3 py-2 text-sm font-medium text-amber-950">
              Evidentiran povrat: {formatMoney(invoice.refund.amount)}
            </p>
          ) : lastCancel?.refund ? (
            <p className="mt-4 rounded-xl border border-amber-200 bg-white/70 px-3 py-2 text-sm font-medium text-amber-950">
              Evidentiran povrat: {formatMoney(lastCancel.refund.amount)}
            </p>
          ) : lastCancel && lastCancel.refundedAmount > 0 ? (
            <p className="mt-4 rounded-xl border border-amber-200 bg-white/70 px-3 py-2 text-sm font-medium text-amber-950">
              Evidentiran povrat: {formatMoney(lastCancel.refundedAmount)}
            </p>
          ) : null}
        </section>
      ) : null}

      <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <h2 className="text-base font-semibold text-stone-900">Stavke</h2>
        <ul className="mt-3 divide-y divide-stone-100">
          {(invoice.items ?? []).map((item) => (
            <li
              key={`${item.productID}-${item.productName}-${item.quantity}`}
              className="flex flex-wrap items-start justify-between gap-2 py-3 first:pt-0 last:pb-0"
            >
              <div className="min-w-0">
                <p className="break-words font-medium text-stone-900">
                  {item.productName}
                </p>
                <p className="mt-1 text-xs text-stone-500">
                  {item.quantity}
                  {item.unit ? ` ${item.unit}` : ''} ×{' '}
                  {formatMoney(item.unitPrice)}
                </p>
              </div>
              <p className="text-sm font-semibold tabular-nums text-stone-900">
                {formatMoney(item.totalPrice)}
              </p>
            </li>
          ))}
        </ul>
        {(invoice.items ?? []).length === 0 ? (
          <p className="mt-2 text-sm text-stone-500">Nema stavki.</p>
        ) : null}
        <p className="mt-3 text-xs text-stone-400">
          Cijene su sačuvane sa računa i ne mijenjaju se naknadno.
        </p>
      </section>

      <CancelInvoiceDialog
        open={cancelOpen}
        invoiceId={invoice.id}
        loading={cancelLoading}
        error={cancelError}
        onClose={() => {
          if (!cancelLoading) {
            setCancelOpen(false);
          }
        }}
        onConfirm={(reason) => void handleCancel(reason)}
      />
    </div>
  );
}
