"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { CustomerOpenInvoices } from "@/components/customers/CustomerOpenInvoices";
import { CustomerPayments } from "@/components/customers/CustomerPayments";
import {
  CustomerStatusBadge,
  DebtBadge,
} from "@/components/customers/CustomerStatusBadge";
import { canViewFinance } from "@/components/dashboard/DashboardHeader";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import {
  deleteCustomer,
  fetchCustomer,
  fetchCustomerFinancialSummary,
  getApiBusinessMessage,
  updateCustomerStatus,
} from "@/lib/customers-api";
import { formatMoney } from "@/lib/format";
import {
  CustomerDetails,
  CustomerFinancialSummary,
} from "@/types/customer";

export function CustomerDetailsView({ customerId }: { customerId: number }) {
  const router = useRouter();
  const { user } = useAuth();
  const showFinance = user ? canViewFinance(user.role) : false;

  const [customer, setCustomer] = useState<CustomerDetails | null>(null);
  const [summary, setSummary] = useState<CustomerFinancialSummary | null>(null);
  const [summaryError, setSummaryError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [confirm, setConfirm] = useState<
    | { open: false }
    | { open: true; kind: "activate" | "deactivate" | "delete" }
  >({ open: false });
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      try {
        const data = await fetchCustomer(customerId);
        if (cancelled) {
          return;
        }
        setCustomer(data);
        setError(null);
      } catch (err) {
        if (cancelled) {
          return;
        }
        setCustomer(null);
        setError(getApiBusinessMessage(err, "Kupac nije pronađen."));
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [customerId, reloadToken]);

  useEffect(() => {
    if (!showFinance) {
      return;
    }
    let cancelled = false;

    async function run() {
      try {
        const data = await fetchCustomerFinancialSummary(customerId);
        if (cancelled) {
          return;
        }
        setSummary(data);
        setSummaryError(null);
      } catch (err) {
        if (cancelled) {
          return;
        }
        setSummary(null);
        setSummaryError(
          getApiBusinessMessage(err, "Finansijski pregled nije dostupan."),
        );
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [customerId, showFinance, reloadToken]);

  async function handleConfirm() {
    if (!confirm.open || !customer) {
      return;
    }
    setConfirmLoading(true);
    setConfirmError(null);
    try {
      if (confirm.kind === "delete") {
        await deleteCustomer(customer.id);
        setConfirm({ open: false });
        router.replace("/customers");
        return;
      }
      await updateCustomerStatus(
        customer.id,
        confirm.kind === "activate",
      );
      setConfirm({ open: false });
      setLoading(true);
      setReloadToken((value) => value + 1);
    } catch (err) {
      setConfirmError(
        getApiBusinessMessage(
          err,
          confirm.kind === "delete"
            ? "Kupac se ne može obrisati jer postoji istorija računa ili uplata. Možete ga deaktivirati ako nema otvorenih računa."
            : "Promjena statusa nije uspjela.",
        ),
      );
    } finally {
      setConfirmLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <ListSkeleton rows={2} />
        <ListSkeleton rows={3} />
      </div>
    );
  }

  if (error || !customer) {
    return (
      <div className="space-y-4">
        <InlineError
          message={error ?? "Kupac nije pronađen."}
          onRetry={() => {
            setLoading(true);
            setReloadToken((value) => value + 1);
          }}
        />
        <Link
          href="/customers"
          className="inline-flex text-sm font-medium text-[#8a6a45]"
        >
          Nazad na listu
        </Link>
      </div>
    );
  }

  const debt = summary?.totalDebt ?? customer.debt;
  const muted = !customer.isActive;

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            Kupac
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-2">
            <h1
              className={`break-words text-2xl font-semibold tracking-tight sm:text-3xl ${
                muted ? "text-stone-600" : "text-stone-900"
              }`}
            >
              {customer.name}
            </h1>
            <CustomerStatusBadge isActive={customer.isActive} />
          </div>
          <p className="mt-1 break-words text-sm text-stone-500">
            {customer.phone?.trim() ? customer.phone : "Bez telefona"}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {customer.isActive ? (
            <Link
              href={`/invoices/new?customerID=${customer.id}`}
              className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800"
            >
              Novi račun za kupca
            </Link>
          ) : null}
          {customer.isActive && debt > 0 ? (
            <Link
              href={`/payments/new?customerID=${customer.id}`}
              className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
            >
              Evidentiraj uplatu
            </Link>
          ) : null}
          <Link
            href={`/customers/${customer.id}/edit`}
            className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
          >
            Uredi
          </Link>
        </div>
      </header>

      <section
        className={`dash-enter rounded-2xl border p-4 sm:p-5 ${
          muted
            ? "border-stone-200/80 bg-stone-50 opacity-90"
            : "border-stone-200 bg-white"
        }`}
      >
        <div className="flex flex-wrap items-center gap-2">
          <DebtBadge amount={debt} />
          <CustomerStatusBadge isActive={customer.isActive} />
        </div>
        <p className="mt-4 text-xs font-medium uppercase tracking-[0.12em] text-stone-500">
          Trenutni dug
        </p>
        <p className="mt-1 text-2xl font-semibold text-stone-950">
          {formatMoney(debt)}
        </p>
        {showFinance && summary ? (
          <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div className="rounded-xl border border-stone-100 bg-stone-50 px-3 py-3">
              <p className="text-xs text-stone-500">Otvoreni računi</p>
              <p className="mt-1 text-lg font-semibold text-stone-900">
                {summary.openInvoicesCount}
              </p>
            </div>
            <div className="rounded-xl border border-stone-100 bg-stone-50 px-3 py-3">
              <p className="text-xs text-stone-500">Broj uplata</p>
              <p className="mt-1 text-lg font-semibold text-stone-900">
                {summary.paymentsCount}
              </p>
            </div>
          </div>
        ) : null}
        {showFinance && summaryError ? (
          <div className="mt-3">
            <InlineError
              message={summaryError}
              onRetry={() => setReloadToken((value) => value + 1)}
            />
          </div>
        ) : null}

        <div className="mt-5 flex flex-wrap gap-2">
          {customer.isActive ? (
            <button
              type="button"
              onClick={() => {
                setConfirmError(null);
                setConfirm({ open: true, kind: "deactivate" });
              }}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
            >
              Deaktiviraj
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                setConfirmError(null);
                setConfirm({ open: true, kind: "activate" });
              }}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
            >
              Aktiviraj
            </button>
          )}
          <button
            type="button"
            onClick={() => {
              setConfirmError(null);
              setConfirm({ open: true, kind: "delete" });
            }}
            className="inline-flex min-h-10 items-center rounded-xl border border-red-200 px-3 text-sm font-medium text-red-700 hover:bg-red-50"
          >
            Obriši
          </button>
        </div>
      </section>

      {customer.invoices?.length ? (
        <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
          <h2 className="text-base font-semibold text-stone-900">
            Računi (sa detalja kupca)
          </h2>
          <ul className="mt-3 space-y-2">
            {customer.invoices.slice(0, 8).map((invoice) => (
              <li
                key={invoice.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-stone-100 px-3 py-2 text-sm"
              >
                <Link
                  href={`/invoices/${invoice.id}`}
                  className="font-medium text-stone-900 hover:text-[#8a6a45]"
                >
                  #{invoice.id}
                </Link>
                <span>{formatMoney(invoice.totalAmount)}</span>
                <span className="text-xs text-stone-500">{invoice.status}</span>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      <CustomerOpenInvoices customerId={customer.id} />
      <CustomerPayments customerId={customer.id} />

      <ConfirmDialog
        open={confirm.open}
        title={
          confirm.open && confirm.kind === "delete"
            ? "Obriši kupca"
            : confirm.open && confirm.kind === "activate"
              ? "Aktiviraj kupca"
              : "Deaktiviraj kupca"
        }
        message={
          confirm.open && confirm.kind === "delete"
            ? `Obrisati kupca „${customer.name}”? Brisanje nije moguće ako postoji istorija računa ili uplata.`
            : confirm.open && confirm.kind === "activate"
              ? `Aktivirati kupca „${customer.name}”?`
              : `Deaktivirati kupca „${customer.name}”? Nije dozvoljeno ako ima unpaid/partially_paid račune.`
        }
        confirmLabel={
          confirm.open && confirm.kind === "delete"
            ? "Obriši"
            : confirm.open && confirm.kind === "activate"
              ? "Aktiviraj"
              : "Deaktiviraj"
        }
        tone={confirm.open && confirm.kind === "delete" ? "danger" : "neutral"}
        loading={confirmLoading}
        error={confirmError}
        onClose={() => {
          if (!confirmLoading) {
            setConfirm({ open: false });
          }
        }}
        onConfirm={() => void handleConfirm()}
      />
    </div>
  );
}
