"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import { formatMoney, formatQuantity, formatUnit } from "@/lib/format";
import {
  fetchSalesReturn,
  getApiBusinessMessage,
} from "@/lib/sales-return-api";
import { userDisplayName } from "@/lib/user-display";
import { SalesReturnDetail, SalesReturnItem } from "@/types/sales-return";

function saleModeLabel(item: SalesReturnItem): string {
  if (item.saleByPackage) {
    return "Prodaje se po pakovanju";
  }
  const unit = formatUnit(item.unit);
  if (unit === "kom") {
    return "Prodaje se po komadu";
  }
  return `Prodaje se po ${unit}`;
}

function ReturnItemCard({ item }: { item: SalesReturnItem }) {
  const unit = formatUnit(item.unit);

  return (
    <article className="rounded-2xl border border-stone-200 bg-white p-4">
      <h3 className="font-medium text-stone-900">{item.productName}</h3>
      <p className="mt-1 text-sm text-stone-600">{saleModeLabel(item)}</p>
      {item.saleByPackage && item.packageQuantity > 0 ? (
        <p className="mt-0.5 text-sm text-stone-600">
          1 paket = {formatQuantity(item.packageQuantity)} {unit}
        </p>
      ) : null}

      <dl className="mt-4 space-y-2 text-sm">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <dt className="text-stone-500">
            {item.saleByPackage ? "Količina vraćena" : "Količina"}
          </dt>
          <dd className="tabular-nums text-stone-900">
            {formatQuantity(item.quantity)} {unit}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <dt className="text-stone-500">Cena povrata</dt>
          <dd className="tabular-nums text-stone-900">
            {formatMoney(item.unitPrice)}/{unit}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-2 border-t border-stone-100 pt-2">
          <dt className="font-medium text-stone-700">Ukupno</dt>
          <dd className="font-semibold tabular-nums text-stone-900">
            {formatMoney(item.totalPrice)}
          </dd>
        </div>
      </dl>
    </article>
  );
}

export function SalesReturnDetailWorkspace({
  salesReturnId,
}: {
  salesReturnId: number;
}) {
  const [detail, setDetail] = useState<SalesReturnDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchSalesReturn(salesReturnId);
      setDetail(data);
    } catch (err) {
      setDetail(null);
      setError(getApiBusinessMessage(err, "Povrat robe nije pronađen."));
    } finally {
      setLoading(false);
    }
  }, [salesReturnId]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void load();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [load, reloadToken]);

  if (loading) {
    return (
      <div className="space-y-4">
        <ListSkeleton rows={2} />
        <ListSkeleton rows={4} />
      </div>
    );
  }

  if (error || !detail) {
    return (
      <div className="space-y-4">
        <InlineError
          message={error ?? "Povrat robe nije pronađen."}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
        <Link
          href="/sales-returns"
          className="inline-flex text-sm font-medium text-[#8a6a45] hover:text-stone-900"
        >
          Nazad na listu
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-sm text-stone-500">
            <Link
              href="/sales-returns"
              className="font-medium text-[#8a6a45] hover:text-stone-900"
            >
              Povrati robe
            </Link>
            <span className="mx-1.5 text-stone-300">/</span>
            #{detail.id}
          </p>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight text-stone-900">
            Povrat robe #{detail.id}
          </h1>
          <p className="mt-1 text-sm text-stone-500">{detail.createdAt}</p>
        </div>
        <Link
          href="/sales-returns"
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Nazad
        </Link>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_280px]">
        <div className="space-y-4">
          <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
            <h2 className="text-sm font-semibold uppercase tracking-[0.08em] text-stone-500">
              Pregled
            </h2>
            <dl className="mt-4 space-y-3 text-sm">
              <div>
                <dt className="text-stone-500">Evidentirao</dt>
                <dd className="mt-0.5 font-medium text-stone-900">
                  {userDisplayName(detail.createdByUser)}
                </dd>
              </div>
              <div>
                <dt className="text-stone-500">Opis</dt>
                <dd className="mt-0.5 text-stone-800">
                  {detail.description.trim() || "—"}
                </dd>
              </div>
              <div>
                <dt className="text-stone-500">Status isplate</dt>
                <dd className="mt-0.5 font-medium text-stone-900">
                  {detail.cashRefunded
                    ? "Novac je vraćen kupcu"
                    : "Bez isplate novca"}
                </dd>
              </div>
              {detail.cashRefunded && detail.refundID ? (
                <div>
                  <dt className="text-stone-500">Finansijski povrat</dt>
                  <dd className="mt-0.5 text-stone-700">
                    Finansijski povrat #{detail.refundID}
                  </dd>
                </div>
              ) : null}
            </dl>
          </section>

          <section className="space-y-3">
            <h2 className="text-sm font-semibold uppercase tracking-[0.08em] text-stone-500">
              Stavke
            </h2>
            <div className="grid gap-3">
              {detail.items.map((item) => (
                <ReturnItemCard key={item.id} item={item} />
              ))}
            </div>
          </section>
        </div>

        <aside className="h-fit rounded-2xl border border-stone-200 bg-white p-4 sm:p-5 lg:sticky lg:top-4">
          <p className="text-sm text-stone-500">Ukupna vrednost povrata</p>
          <p className="mt-1 text-2xl font-semibold tabular-nums text-stone-900">
            {formatMoney(detail.totalAmount)}
          </p>

          <div className="mt-4 border-t border-stone-100 pt-4">
            {detail.cashRefunded ? (
              <>
                <p className="text-sm text-stone-500">Isplaćeno kupcu</p>
                <p className="mt-1 text-lg font-semibold tabular-nums text-stone-900">
                  {formatMoney(detail.totalAmount)}
                </p>
              </>
            ) : (
              <>
                <p className="text-sm text-stone-500">Finansijska isplata</p>
                <p className="mt-1 text-sm font-medium text-stone-700">
                  Nije izvršena
                </p>
              </>
            )}
          </div>
        </aside>
      </div>
    </div>
  );
}
