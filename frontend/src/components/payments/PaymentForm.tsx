"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import { CustomerSelector } from "@/components/customers/CustomerSelector";
import { InvoiceStatusBadge } from "@/components/invoices/InvoiceStatusBadge";
import { PaymentInvoiceSuccessPanel } from "@/components/payments/PaymentInvoiceSuccessPanel";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import {
  fetchCustomer,
  fetchCustomerOpenInvoices,
  getApiBusinessMessage,
} from "@/lib/customers-api";
import { formatMoney } from "@/lib/format";
import { fetchInvoice, invoiceCustomerLabel } from "@/lib/invoices-api";
import {
  autoAllocatePayments,
  createPayment,
  roundMoney,
} from "@/lib/payments-api";
import { CustomerListItem, CustomerOpenInvoice } from "@/types/customer";
import { InvoiceDetails } from "@/types/invoice";
import { Payment } from "@/types/payment";

type Mode = "invoice" | "customer";

function moneyInputValue(value: number): string {
  if (!Number.isFinite(value) || value === 0) {
    return "";
  }
  return String(value);
}

export function PaymentForm({
  initialInvoiceID,
  initialCustomerID,
}: {
  initialInvoiceID?: number | null;
  initialCustomerID?: number | null;
}) {
  const router = useRouter();
  const mode: Mode = initialInvoiceID ? "invoice" : "customer";

  const [loadingBootstrap, setLoadingBootstrap] = useState(true);
  const [bootstrapError, setBootstrapError] = useState<string | null>(null);

  const [invoice, setInvoice] = useState<InvoiceDetails | null>(null);
  const [customer, setCustomer] = useState<CustomerListItem | null>(null);
  const [openInvoices, setOpenInvoices] = useState<CustomerOpenInvoice[]>([]);

  const [totalAmount, setTotalAmount] = useState(0);
  const [allocations, setAllocations] = useState<Record<number, number>>({});
  const [amountMode, setAmountMode] = useState<"full" | "custom">("full");

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [staleHint, setStaleHint] = useState(false);
  const [reloadToken, setReloadToken] = useState(0);
  const [successPayment, setSuccessPayment] = useState<Payment | null>(null);
  const [successInvoice, setSuccessInvoice] = useState<InvoiceDetails | null>(
    null,
  );

  const loadInvoiceMode = useCallback(async (invoiceID: number) => {
    const data = await fetchInvoice(invoiceID);
    if (data.customerID == null && !data.customer) {
      throw new Error(
        "Gotovinski račun nema kupca — uplata se evidentira samo za račune kupaca.",
      );
    }
    if (data.status === "cancelled") {
      throw new Error("Ne može se evidentirati uplata na storniran račun.");
    }
    if (data.status === "paid" || data.remainingAmount <= 0) {
      throw new Error("Račun je već u potpunosti plaćen.");
    }
    const customerData = data.customer
      ? {
          id: data.customer.id,
          name: data.customer.name,
          phone: data.customer.phone,
          isActive: data.customer.isActive,
        }
      : await fetchCustomer(data.customerID as number).then((c) => ({
          id: c.id,
          name: c.name,
          phone: c.phone,
          isActive: c.isActive,
        }));
    if (!customerData.isActive) {
      throw new Error("Kupac računa nije aktivan.");
    }
    setInvoice(data);
    setCustomer(customerData);
    setOpenInvoices([]);
    setAmountMode("full");
    setTotalAmount(roundMoney(data.remainingAmount));
    setAllocations({ [data.id]: roundMoney(data.remainingAmount) });
  }, []);

  const loadCustomerMode = useCallback(async (customerID: number | null) => {
    if (!customerID) {
      setCustomer(null);
      setOpenInvoices([]);
      setAllocations({});
      setTotalAmount(0);
      return;
    }
    const details = await fetchCustomer(customerID);
    if (!details.isActive) {
      throw new Error("Kupac nije aktivan.");
    }
    const opens = await fetchCustomerOpenInvoices(customerID);
    setCustomer({
      id: details.id,
      name: details.name,
      phone: details.phone,
      isActive: details.isActive,
    });
    setOpenInvoices(opens);
    setInvoice(null);
    setAllocations({});
    setTotalAmount(0);
  }, []);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      setLoadingBootstrap(true);
      setBootstrapError(null);
      setError(null);
      setStaleHint(false);
      try {
        if (mode === "invoice" && initialInvoiceID) {
          await loadInvoiceMode(initialInvoiceID);
        } else if (initialCustomerID) {
          await loadCustomerMode(initialCustomerID);
        } else {
          setCustomer(null);
          setOpenInvoices([]);
          setInvoice(null);
          setAllocations({});
          setTotalAmount(0);
        }
      } catch (err) {
        if (!cancelled) {
          setBootstrapError(
            getApiBusinessMessage(err, "Podaci za uplatu nisu učitani."),
          );
        }
      } finally {
        if (!cancelled) {
          setLoadingBootstrap(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [
    mode,
    initialInvoiceID,
    initialCustomerID,
    loadInvoiceMode,
    loadCustomerMode,
    reloadToken,
  ]);

  const allocatedSum = useMemo(() => {
    return roundMoney(
      Object.values(allocations).reduce(
        (sum, value) => sum + (Number.isFinite(value) ? value : 0),
        0,
      ),
    );
  }, [allocations]);

  const unallocated = roundMoney(totalAmount - allocatedSum);
  const coveredCount = Object.values(allocations).filter((v) => v > 0).length;

  const validationError = useMemo(() => {
    if (!customer) {
      return "Izaberite aktivnog kupca.";
    }
    if (!(totalAmount > 0)) {
      return "Primljeni iznos mora biti veći od 0.";
    }
    if (coveredCount === 0) {
      return "Dodajte najmanje jednu raspodelu.";
    }
    if (Math.abs(unallocated) > 0.009) {
      return "Zbir raspodele mora tačno odgovarati primljenom iznosu.";
    }
    if (mode === "invoice" && invoice) {
      const amount = allocations[invoice.id] ?? 0;
      if (amount > invoice.remainingAmount + 0.009) {
        return "Iznos ne sme prelaziti preostalo na računu.";
      }
    }
    for (const open of openInvoices) {
      const amount = allocations[open.id] ?? 0;
      if (amount < 0) {
        return "Iznosi raspodele ne mogu biti negativni.";
      }
      if (amount > open.remainingAmount + 0.009) {
        return `Raspodela za račun #${open.id} prelazi preostalo.`;
      }
    }
    return null;
  }, [
    allocations,
    coveredCount,
    customer,
    invoice,
    mode,
    openInvoices,
    totalAmount,
    unallocated,
  ]);

  const canSubmit =
    !validationError &&
    !submitting &&
    !loadingBootstrap &&
    !successPayment;

  function setFullInvoiceAmount() {
    if (!invoice) {
      return;
    }
    setAmountMode("full");
    const remaining = roundMoney(invoice.remainingAmount);
    setTotalAmount(remaining);
    setAllocations({ [invoice.id]: remaining });
  }

  function handleAutoAllocate() {
    if (totalAmount <= 0 || openInvoices.length === 0) {
      return;
    }
    const next = autoAllocatePayments(totalAmount, openInvoices);
    const map: Record<number, number> = {};
    for (const [id, amount] of next.entries()) {
      map[id] = amount;
    }
    setAllocations(map);
  }

  async function handleSubmit() {
    if (!canSubmit || !customer || validationError || successPayment) {
      setError(validationError);
      return;
    }
    setSubmitting(true);
    setError(null);
    setStaleHint(false);
    try {
      const payloadAllocations = Object.entries(allocations)
        .map(([invoiceID, amount]) => ({
          invoiceID: Number(invoiceID),
          amount: roundMoney(amount),
        }))
        .filter((item) => item.amount > 0);

      const payment = await createPayment({
        customerID: customer.id,
        totalAmount: roundMoney(totalAmount),
        allocations: payloadAllocations,
      });

      if (mode === "invoice" && invoice) {
        const updated = await fetchInvoice(invoice.id);
        setSuccessPayment(payment);
        setSuccessInvoice(updated);
      } else {
        router.replace(`/payments/${payment.id}`);
      }
    } catch (err) {
      const message = getApiBusinessMessage(
        err,
        "Evidentiranje uplate nije uspelo.",
      );
      setError(message);
      const lower = message.toLowerCase();
      if (
        lower.includes("preostal") ||
        lower.includes("vec placen") ||
        lower.includes("veći od") ||
        lower.includes("veci od")
      ) {
        setStaleHint(true);
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (loadingBootstrap) {
    return (
      <div className="space-y-4">
        <ListSkeleton rows={3} />
      </div>
    );
  }

  if (bootstrapError) {
    return (
      <div className="space-y-4">
        <InlineError
          message={bootstrapError}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
        <Link href="/payments" className="text-sm font-medium text-[#8a6a45]">
          Lista uplata
        </Link>
      </div>
    );
  }

  return (
    <div className="min-w-0 pb-28 lg:pb-0">
      <header className="mb-4">
        <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
          Nova uplata
        </h1>
        <p className="mt-1 text-sm text-stone-500">
          {mode === "invoice"
            ? "Brza uplata za jedan račun."
            : "Uplata kupca sa raspodelom na otvorene račune."}
        </p>
      </header>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.6fr)_minmax(17rem,0.9fr)] lg:items-start">
        <div className="min-w-0 space-y-4">
          {mode === "invoice" && invoice ? (
            <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="text-base font-semibold text-stone-900">
                  Račun #{invoice.id}
                </h2>
                <InvoiceStatusBadge status={invoice.status} />
              </div>
              <dl className="mt-3 grid grid-cols-2 gap-3 text-sm sm:grid-cols-3">
                <div>
                  <dt className="text-stone-500">Kupac</dt>
                  <dd className="font-medium text-stone-900">
                    {customer?.name ?? "—"}
                  </dd>
                </div>
                <div>
                  <dt className="text-stone-500">Ukupno</dt>
                  <dd className="tabular-nums text-stone-900">
                    {formatMoney(invoice.totalAmount)}
                  </dd>
                </div>
                <div>
                  <dt className="text-stone-500">Plaćeno</dt>
                  <dd className="tabular-nums text-stone-900">
                    {formatMoney(invoice.paidAmount)}
                  </dd>
                </div>
                <div>
                  <dt className="text-stone-500">Preostalo</dt>
                  <dd className="font-semibold tabular-nums text-stone-900">
                    {formatMoney(invoice.remainingAmount)}
                  </dd>
                </div>
                <div>
                  <dt className="text-stone-500">Datum</dt>
                  <dd className="text-stone-800">{invoice.createdAt}</dd>
                </div>
              </dl>

              <div className="mt-4 flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={setFullInvoiceAmount}
                  className={`rounded-xl px-3 py-2 text-sm font-medium ${
                    amountMode === "full"
                      ? "bg-stone-900 text-white"
                      : "border border-stone-200 bg-white text-stone-700"
                  }`}
                >
                  Plati sve
                </button>
                <button
                  type="button"
                  onClick={() => setAmountMode("custom")}
                  className={`rounded-xl px-3 py-2 text-sm font-medium ${
                    amountMode === "custom"
                      ? "bg-stone-900 text-white"
                      : "border border-stone-200 bg-white text-stone-700"
                  }`}
                >
                  Drugi iznos
                </button>
              </div>

              {amountMode === "custom" ? (
                <label className="mt-3 block text-sm">
                  <span className="mb-1.5 block font-medium text-stone-700">
                    Iznos uplate (RSD)
                  </span>
                  <input
                    type="number"
                    inputMode="decimal"
                    step="0.01"
                    min="0.01"
                    max={invoice.remainingAmount}
                    value={moneyInputValue(totalAmount)}
                    onChange={(event) => {
                      const next = roundMoney(Number(event.target.value));
                      setTotalAmount(next);
                      setAllocations({
                        [invoice.id]: next > 0 ? next : 0,
                      });
                    }}
                    className="w-full max-w-xs rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
                  />
                </label>
              ) : (
                <p className="mt-3 text-sm text-stone-600">
                  Iznos:{" "}
                  <span className="font-semibold tabular-nums text-stone-900">
                    {formatMoney(totalAmount)}
                  </span>
                </p>
              )}
            </section>
          ) : (
            <>
              <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
                <CustomerSelector
                  value={customer}
                  onChange={(next) => {
                    setError(null);
                    if (next) {
                      router.replace(`/payments/new?customerID=${next.id}`);
                    } else {
                      setCustomer(null);
                      setOpenInvoices([]);
                      setAllocations({});
                      setTotalAmount(0);
                      router.replace("/payments/new");
                    }
                  }}
                />
              </section>

              {customer ? (
                <>
                  <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
                    <div className="flex flex-wrap items-end justify-between gap-3">
                      <label className="block min-w-0 flex-1 text-sm">
                        <span className="mb-1.5 block font-medium text-stone-700">
                          Primljeni iznos (RSD)
                        </span>
                        <input
                          type="number"
                          inputMode="decimal"
                          step="0.01"
                          min="0.01"
                          value={moneyInputValue(totalAmount)}
                          onChange={(event) => {
                            setTotalAmount(
                              roundMoney(Number(event.target.value)),
                            );
                          }}
                          className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
                        />
                      </label>
                      <button
                        type="button"
                        onClick={handleAutoAllocate}
                        className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-800 hover:bg-stone-50"
                      >
                        Automatski rasporedi
                      </button>
                    </div>
                  </section>

                  <section className="space-y-2">
                    <h2 className="text-sm font-semibold text-stone-800">
                      Otvoreni računi
                    </h2>
                    {openInvoices.length === 0 ? (
                      <p className="rounded-2xl border border-dashed border-stone-200 bg-white px-4 py-8 text-center text-sm text-stone-500">
                        Kupac nema otvorenih računa.
                      </p>
                    ) : (
                      <ul className="space-y-2">
                        {openInvoices.map((open) => (
                          <li
                            key={open.id}
                            className="rounded-2xl border border-stone-200 bg-white p-3 sm:p-4"
                          >
                            <div className="flex flex-wrap items-start justify-between gap-2">
                              <div className="min-w-0">
                                <div className="flex flex-wrap items-center gap-2">
                                  <Link
                                    href={`/invoices/${open.id}`}
                                    className="font-medium text-stone-900 hover:text-[#8a6a45]"
                                  >
                                    Račun #{open.id}
                                  </Link>
                                  <InvoiceStatusBadge status={open.status} />
                                </div>
                                <p className="mt-1 text-xs text-stone-500">
                                  {open.createdAt} · preostalo{" "}
                                  {formatMoney(open.remainingAmount)}
                                </p>
                              </div>
                              <label className="block text-sm">
                                <span className="sr-only">
                                  Iznos za račun {open.id}
                                </span>
                                <input
                                  type="number"
                                  inputMode="decimal"
                                  step="0.01"
                                  min="0"
                                  max={open.remainingAmount}
                                  value={moneyInputValue(
                                    allocations[open.id] ?? 0,
                                  )}
                                  onChange={(event) => {
                                    const next = roundMoney(
                                      Number(event.target.value),
                                    );
                                    setAllocations((current) => ({
                                      ...current,
                                      [open.id]: next > 0 ? next : 0,
                                    }));
                                  }}
                                  className="w-28 rounded-xl border border-stone-200 px-3 py-2 text-sm tabular-nums outline-none ring-[#c4a484]/40 focus:ring-2"
                                />
                              </label>
                            </div>
                          </li>
                        ))}
                      </ul>
                    )}
                  </section>
                </>
              ) : null}
            </>
          )}
        </div>

        <aside className="hidden lg:block">
          <PaymentSummaryPanel
            totalAmount={totalAmount}
            allocatedSum={allocatedSum}
            unallocated={unallocated}
            coveredCount={coveredCount}
            customerName={customer?.name ?? null}
            error={error ?? validationError}
            staleHint={staleHint}
            submitting={submitting}
            canSubmit={canSubmit}
            primaryLabel="Evidentiraj uplatu"
            onSubmit={() => void handleSubmit()}
            onRefresh={() => setReloadToken((value) => value + 1)}
          />
        </aside>
      </div>

      <div className="fixed inset-x-0 bottom-0 z-30 border-t border-stone-200 bg-white/95 px-4 py-3 backdrop-blur lg:hidden">
        <div className="flex w-full items-center gap-3">
          <div className="min-w-0 flex-1 text-sm">
            <p className="text-stone-500">
              Primljeno {formatMoney(totalAmount)}
            </p>
            <p className="font-semibold tabular-nums text-stone-900">
              Raspoređeno {formatMoney(allocatedSum)}
            </p>
          </div>
          <button
            type="button"
            disabled={!canSubmit}
            onClick={() => void handleSubmit()}
            className="inline-flex min-h-11 shrink-0 items-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white disabled:opacity-50"
          >
            {submitting ? "Slanje…" : "Evidentiraj uplatu"}
          </button>
        </div>
        {(error || (validationError && totalAmount > 0)) && (
          <p className="mt-2 w-full break-words text-xs text-red-700">
            {error ?? validationError}
          </p>
        )}
      </div>

      {successPayment && successInvoice ? (
        <PaymentInvoiceSuccessPanel
          invoice={successInvoice}
          payment={successPayment}
          customerLabel={invoiceCustomerLabel(successInvoice)}
        />
      ) : null}
    </div>
  );
}

function PaymentSummaryPanel({
  totalAmount,
  allocatedSum,
  unallocated,
  coveredCount,
  customerName,
  error,
  staleHint,
  submitting,
  canSubmit,
  primaryLabel,
  onSubmit,
  onRefresh,
}: {
  totalAmount: number;
  allocatedSum: number;
  unallocated: number;
  coveredCount: number;
  customerName: string | null;
  error: string | null;
  staleHint: boolean;
  submitting: boolean;
  canSubmit: boolean;
  primaryLabel: string;
  onSubmit: () => void;
  onRefresh: () => void;
}) {
  return (
    <div className="sticky top-4 rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <h2 className="text-base font-semibold text-stone-900">Pregled uplate</h2>
      <p className="mt-1 text-sm text-stone-500">
        {customerName ?? "Kupac nije izabran"}
      </p>
      <dl className="mt-4 space-y-2 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Primljeno</dt>
          <dd className="font-medium tabular-nums text-stone-900">
            {formatMoney(totalAmount)}
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Raspoređeno</dt>
          <dd className="font-medium tabular-nums text-stone-900">
            {formatMoney(allocatedSum)}
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-stone-500">Neraspoređeno</dt>
          <dd
            className={`font-medium tabular-nums ${
              Math.abs(unallocated) > 0.009
                ? "text-amber-800"
                : "text-stone-900"
            }`}
          >
            {formatMoney(unallocated)}
          </dd>
        </div>
        <div className="flex justify-between gap-3 border-t border-stone-100 pt-2">
          <dt className="text-stone-500">Računi</dt>
          <dd className="font-medium text-stone-900">{coveredCount}</dd>
        </div>
      </dl>

      {error ? (
        <p className="mt-3 break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      ) : null}
      {staleHint ? (
        <button
          type="button"
          onClick={onRefresh}
          className="mt-2 w-full rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm font-medium text-amber-900"
        >
          Osveži podatke
        </button>
      ) : null}

      <button
        type="button"
        disabled={!canSubmit}
        onClick={onSubmit}
        className="mt-4 inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {submitting ? "Slanje…" : primaryLabel}
      </button>
    </div>
  );
}
