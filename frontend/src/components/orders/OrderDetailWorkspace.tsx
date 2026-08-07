"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import { CustomerSelector } from "@/components/customers/CustomerSelector";
import { InvoiceDocumentActions } from "@/components/invoices/InvoiceDocumentActions";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import { formatMoney, formatQuantity } from "@/lib/format";
import {
  confirmOnlineOrder,
  deleteOnlineOrder,
  fetchOnlineOrderById,
  formatOrderDateTime,
  formatRelativeReceived,
  getApiBusinessMessage,
  notifyOnlineOrdersChanged,
  onlineOrderCustomerName,
  onlineOrderStatusLabel,
} from "@/lib/online-orders-api";
import { CustomerListItem } from "@/types/customer";
import {
  OnlineOrderDetail,
  OnlineOrderItemDetail,
} from "@/types/online-order-staff";

function statusBadgeClass(status: string): string {
  if (status === "pending") {
    return "bg-amber-50 text-amber-900 ring-amber-200";
  }
  if (status === "confirmed") {
    return "bg-emerald-50 text-emerald-800 ring-emerald-200";
  }
  return "bg-stone-100 text-stone-700 ring-stone-200";
}

function itemNeedsAttention(item: OnlineOrderItemDetail): boolean {
  return (
    item.currentInStockEnough === false || item.currentProductActive === false
  );
}

function itemWarningText(item: OnlineOrderItemDetail): string | null {
  const parts: string[] = [];
  if (item.currentProductActive === false) {
    parts.push("proizvod nije aktivan");
  }
  if (item.currentInStockEnough === false) {
    parts.push("nema dovoljno na lageru");
  }
  return parts.length ? parts.join(" · ") : null;
}

type CustomerMode = "existing" | "new";

export function OrderDetailWorkspace({ orderId }: { orderId: number }) {
  const router = useRouter();

  const [order, setOrder] = useState<OnlineOrderDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const [customerMode, setCustomerMode] = useState<CustomerMode>("existing");
  const [selectedCustomer, setSelectedCustomer] =
    useState<CustomerListItem | null>(null);
  const [newCustomerName, setNewCustomerName] = useState("");
  const [newCustomerPhone, setNewCustomerPhone] = useState("");

  const [confirmLoading, setConfirmLoading] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);
  const [confirmedInvoiceId, setConfirmedInvoiceId] = useState<number | null>(
    null,
  );

  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const loadOrder = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchOnlineOrderById(orderId);
      setOrder(data);
      setNewCustomerName(
        `${data.firstName} ${data.lastName}`.trim(),
      );
      setNewCustomerPhone(data.phone ?? "");
      if (data.status === "confirmed" && data.invoiceID) {
        setConfirmedInvoiceId(data.invoiceID);
      }
    } catch (err) {
      setOrder(null);
      setError(getApiBusinessMessage(err, "Narudžbina nije pronađena."));
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadOrder();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadOrder, reloadToken]);

  const customerName = useMemo(
    () => (order ? onlineOrderCustomerName(order) : ""),
    [order],
  );

  const canConfirm =
    order?.status === "pending" &&
    (customerMode === "existing"
      ? selectedCustomer != null
      : newCustomerName.trim().length > 0 && newCustomerPhone.trim().length > 0);

  async function handleConfirm() {
    if (!order || confirmLoading || !canConfirm) {
      return;
    }
    setConfirmLoading(true);
    setConfirmError(null);
    try {
      const body =
        customerMode === "existing"
          ? { customerID: selectedCustomer!.id }
          : {
              newCustomer: {
                name: newCustomerName.trim(),
                phone: newCustomerPhone.trim(),
              },
            };
      const result = await confirmOnlineOrder(order.id, body);
      setConfirmedInvoiceId(result.invoiceID);
      notifyOnlineOrdersChanged();
      setReloadToken((value) => value + 1);
    } catch (err) {
      setConfirmError(
        getApiBusinessMessage(err, "Potvrda narudžbine nije uspela."),
      );
    } finally {
      setConfirmLoading(false);
    }
  }

  async function handleDelete() {
    if (!order || deleteLoading) {
      return;
    }
    setDeleteLoading(true);
    setDeleteError(null);
    try {
      await deleteOnlineOrder(order.id);
      notifyOnlineOrdersChanged();
      router.replace("/orders");
    } catch (err) {
      setDeleteError(
        getApiBusinessMessage(err, "Brisanje narudžbine nije uspelo."),
      );
      setDeleteLoading(false);
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

  if (error || !order) {
    return (
      <div className="space-y-4">
        <InlineError
          message={error ?? "Narudžbina nije pronađena."}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
        <Link href="/orders" className="text-sm font-medium text-[#8a6a45]">
          Nazad na narudžbine
        </Link>
      </div>
    );
  }

  const isPending = order.status === "pending";
  const invoiceId = confirmedInvoiceId ?? order.invoiceID ?? null;
  const showConfirmForm = isPending && invoiceId == null;
  const isConfirmed = order.status === "confirmed" || invoiceId != null;

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            Online narudžbina
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
              Narudžbina #{order.id}
            </h1>
            <span
              className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${statusBadgeClass(order.status)}`}
            >
              {onlineOrderStatusLabel(order.status)}
            </span>
          </div>
          <p className="mt-1 text-sm text-stone-500">
            Primljena {formatRelativeReceived(order.createdAt)} ·{" "}
            {formatOrderDateTime(order.createdAt)}
          </p>
        </div>
        <Link
          href="/orders"
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Nazad na listu
        </Link>
      </header>

      {isConfirmed && invoiceId ? (
        <section className="dash-enter rounded-2xl border border-emerald-200 bg-emerald-50/60 p-4 sm:p-5">
          <p className="text-sm font-medium text-emerald-900">
            Narudžbina je potvrđena. Kreiran je račun #{invoiceId}.
          </p>
          <div className="mt-3">
            <InvoiceDocumentActions
              invoiceId={invoiceId}
              variant="inline"
              showOpen
              openLabel="Otvori račun"
              printLabel="Štampaj"
            />
          </div>
          <Link
            href={`/invoices/${invoiceId}`}
            className="mt-3 inline-flex text-sm font-medium text-[#8a6a45] hover:underline"
          >
            Idi na račun #{invoiceId}
          </Link>
        </section>
      ) : null}

      <section className="dash-enter rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <h2 className="text-sm font-semibold text-stone-900">Kupac</h2>
        <dl className="mt-3 grid gap-3 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-stone-500">Ime</dt>
            <dd className="mt-0.5 font-medium text-stone-900">{customerName}</dd>
          </div>
          <div>
            <dt className="text-stone-500">Telefon</dt>
            <dd className="mt-0.5 flex flex-wrap items-center gap-2">
              <span className="font-medium text-stone-900">{order.phone}</span>
              {order.phone ? (
                <a
                  href={`tel:${order.phone.replace(/\s+/g, "")}`}
                  className="inline-flex min-h-9 items-center rounded-lg border border-stone-200 px-2.5 text-xs font-medium text-stone-700 hover:bg-stone-50"
                >
                  Pozovi
                </a>
              ) : null}
            </dd>
          </div>
          <div>
            <dt className="text-stone-500">Grad</dt>
            <dd className="mt-0.5 text-stone-800">{order.city || "—"}</dd>
          </div>
          <div>
            <dt className="text-stone-500">Adresa</dt>
            <dd className="mt-0.5 break-words text-stone-800">
              {order.address || "—"}
            </dd>
          </div>
          {order.email ? (
            <div className="sm:col-span-2">
              <dt className="text-stone-500">Email</dt>
              <dd className="mt-0.5 break-words text-stone-800">{order.email}</dd>
            </div>
          ) : null}
        </dl>
        {order.note?.trim() ? (
          <div className="mt-4 rounded-xl border border-stone-100 bg-stone-50 px-3 py-3">
            <p className="text-xs font-medium uppercase tracking-wide text-stone-500">
              Napomena
            </p>
            <p className="mt-1 break-words text-sm text-stone-800">
              {order.note}
            </p>
          </div>
        ) : null}
      </section>

      <section className="dash-enter overflow-hidden rounded-2xl border border-stone-200 bg-white">
        <div className="border-b border-stone-100 px-4 py-3 sm:px-5">
          <h2 className="text-sm font-semibold text-stone-900">Proizvodi</h2>
        </div>

        <ul className="divide-y divide-stone-100 lg:hidden">
          {order.items.map((item) => {
            const warning = itemWarningText(item);
            const attention = itemNeedsAttention(item);
            return (
              <li
                key={`${item.productID}-${item.productName}`}
                className={`px-4 py-3 ${attention ? "bg-amber-50/70" : ""}`}
              >
                <p className="font-medium text-stone-900">{item.productName}</p>
                <p className="mt-1 text-sm text-stone-600">
                  {formatQuantity(item.quantity)} {item.unit} ·{" "}
                  {formatMoney(item.unitPrice)}
                </p>
                <p className="mt-1 font-medium tabular-nums text-stone-900">
                  {formatMoney(item.totalPrice)}
                </p>
                {warning ? (
                  <p className="mt-1 text-xs font-medium text-amber-800">
                    {warning}
                  </p>
                ) : null}
              </li>
            );
          })}
        </ul>

        <div className="hidden overflow-x-auto lg:block">
          <table className="w-full text-left text-sm">
            <thead className="bg-stone-50/95">
              <tr className="border-b border-stone-200 text-xs uppercase tracking-[0.08em] text-stone-500">
                <th className="px-4 py-3 font-medium">Proizvod</th>
                <th className="px-4 py-3 font-medium text-right">Količina</th>
                <th className="px-4 py-3 font-medium text-right">Cena</th>
                <th className="px-4 py-3 font-medium text-right">Ukupno</th>
              </tr>
            </thead>
            <tbody>
              {order.items.map((item) => {
                const warning = itemWarningText(item);
                const attention = itemNeedsAttention(item);
                return (
                  <tr
                    key={`${item.productID}-${item.productName}`}
                    className={`border-b border-stone-100 last:border-b-0 ${
                      attention ? "bg-amber-50/70" : ""
                    }`}
                  >
                    <td className="px-4 py-3 align-top">
                      <p className="font-medium text-stone-900">
                        {item.productName}
                      </p>
                      {warning ? (
                        <p className="mt-1 text-xs font-medium text-amber-800">
                          {warning}
                        </p>
                      ) : null}
                    </td>
                    <td className="px-4 py-3 align-top text-right tabular-nums text-stone-700">
                      {formatQuantity(item.quantity)} {item.unit}
                    </td>
                    <td className="px-4 py-3 align-top text-right tabular-nums text-stone-700">
                      {formatMoney(item.unitPrice)}
                    </td>
                    <td className="px-4 py-3 align-top text-right font-medium tabular-nums text-stone-900">
                      {formatMoney(item.totalPrice)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div className="border-t border-stone-100 px-4 py-4 sm:px-5">
          <div className="flex flex-wrap items-end justify-between gap-2">
            <div>
              <p className="text-sm text-stone-500">Vrednost proizvoda</p>
              <p className="text-lg font-semibold tabular-nums text-stone-950">
                {formatMoney(order.totalAmount)}
              </p>
            </div>
          </div>
          <p className="mt-2 text-xs text-stone-500">
            Troškovi transporta nisu uračunati u cenu.
          </p>
        </div>
      </section>

      {showConfirmForm ? (
        <section className="dash-enter space-y-4 rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
          <div>
            <h2 className="text-sm font-semibold text-stone-900">
              Potvrda narudžbine
            </h2>
            <p className="mt-1 text-sm text-stone-500">
              Izaberite postojećeg kupca ili kreirajte novog, zatim potvrdite i
              kreirajte račun.
            </p>
          </div>

          <div
            className="flex flex-wrap gap-1 rounded-xl bg-stone-100 p-1"
            role="tablist"
            aria-label="Način kupca"
          >
            <button
              type="button"
              role="tab"
              aria-selected={customerMode === "existing"}
              onClick={() => setCustomerMode("existing")}
              className={`rounded-lg px-3 py-2 text-sm font-medium transition ${
                customerMode === "existing"
                  ? "bg-white text-stone-900 shadow-sm"
                  : "text-stone-600 hover:text-stone-900"
              }`}
            >
              Postojeći kupac
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={customerMode === "new"}
              onClick={() => setCustomerMode("new")}
              className={`rounded-lg px-3 py-2 text-sm font-medium transition ${
                customerMode === "new"
                  ? "bg-white text-stone-900 shadow-sm"
                  : "text-stone-600 hover:text-stone-900"
              }`}
            >
              Novi kupac
            </button>
          </div>

          {customerMode === "existing" ? (
            <CustomerSelector
              value={selectedCustomer}
              onChange={setSelectedCustomer}
              label="Kupac"
            />
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              <label className="block text-sm">
                <span className="mb-1.5 block font-medium text-stone-700">
                  Ime kupca
                </span>
                <input
                  value={newCustomerName}
                  onChange={(event) => setNewCustomerName(event.target.value)}
                  className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
                />
              </label>
              <label className="block text-sm">
                <span className="mb-1.5 block font-medium text-stone-700">
                  Telefon
                </span>
                <input
                  value={newCustomerPhone}
                  onChange={(event) => setNewCustomerPhone(event.target.value)}
                  className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
                />
              </label>
            </div>
          )}

          <div className="rounded-xl border border-stone-100 bg-stone-50 px-3 py-3 text-sm text-stone-700">
            <p>
              <span className="text-stone-500">Narudžbina:</span> #{order.id}
            </p>
            <p className="mt-1">
              <span className="text-stone-500">Kupac na narudžbini:</span>{" "}
              {customerName} · {order.phone}
            </p>
            <p className="mt-1">
              <span className="text-stone-500">Račun za:</span>{" "}
              {customerMode === "existing"
                ? selectedCustomer
                  ? `${selectedCustomer.name}${
                      selectedCustomer.phone
                        ? ` · ${selectedCustomer.phone}`
                        : ""
                    }`
                  : "Nije izabran"
                : `${newCustomerName.trim() || "—"} · ${
                    newCustomerPhone.trim() || "—"
                  }`}
            </p>
            <p className="mt-1 font-medium tabular-nums text-stone-900">
              Vrednost: {formatMoney(order.totalAmount)}
            </p>
          </div>

          {confirmError ? (
            <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
              {confirmError}
            </p>
          ) : null}

          <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
            <button
              type="button"
              disabled={!canConfirm || confirmLoading}
              onClick={() => void handleConfirm()}
              className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {confirmLoading ? "Sačekajte..." : "Potvrdi i kreiraj račun"}
            </button>
            <button
              type="button"
              disabled={confirmLoading || deleteLoading}
              onClick={() => {
                setDeleteError(null);
                setDeleteOpen(true);
              }}
              className="inline-flex min-h-11 items-center justify-center rounded-xl border border-red-200 px-4 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
            >
              Obriši narudžbinu
            </button>
          </div>
        </section>
      ) : null}

      <ConfirmDialog
        open={deleteOpen}
        title="Obriši narudžbinu"
        message={`Da li ste sigurni da želite da obrišete narudžbinu #${order.id}? Ova akcija se ne može poništiti.`}
        confirmLabel="Obriši"
        cancelLabel="Otkaži"
        loading={deleteLoading}
        error={deleteError}
        tone="danger"
        onConfirm={() => void handleDelete()}
        onClose={() => {
          if (!deleteLoading) {
            setDeleteOpen(false);
          }
        }}
      />
    </div>
  );
}
