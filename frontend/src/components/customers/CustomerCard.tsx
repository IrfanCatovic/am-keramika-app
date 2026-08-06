"use client";

import Link from "next/link";

import { CustomerStatusBadge } from "@/components/customers/CustomerStatusBadge";
import { CustomerListItem } from "@/types/customer";

export function CustomerCard({
  customer,
  busy,
  onActivate,
  onDeactivate,
  onDelete,
}: {
  customer: CustomerListItem;
  busy: boolean;
  onActivate: () => void;
  onDeactivate: () => void;
  onDelete: () => void;
}) {
  const muted = !customer.isActive;

  return (
    <article
      className={`dash-enter min-w-0 rounded-2xl border p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)] ${
        muted
          ? "border-stone-200/80 bg-stone-50 opacity-75"
          : "border-stone-200 bg-white"
      }`}
    >
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <Link
            href={`/customers/${customer.id}`}
            className={`break-words text-base font-semibold hover:text-[#8a6a45] ${
              muted ? "text-stone-600" : "text-stone-900"
            }`}
          >
            {customer.name}
          </Link>
          <p className="mt-1 break-words text-sm text-stone-600">
            {customer.phone?.trim() ? customer.phone : "Bez telefona"}
          </p>
        </div>
        <CustomerStatusBadge isActive={customer.isActive} />
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        <Link
          href={`/customers/${customer.id}`}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Detalji
        </Link>
        <Link
          href={`/customers/${customer.id}/edit`}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Uredi
        </Link>
        {customer.isActive ? (
          <button
            type="button"
            disabled={busy}
            onClick={onDeactivate}
            className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
          >
            Deaktiviraj
          </button>
        ) : (
          <button
            type="button"
            disabled={busy}
            onClick={onActivate}
            className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
          >
            Aktiviraj
          </button>
        )}
        <button
          type="button"
          disabled={busy}
          onClick={onDelete}
          className="inline-flex min-h-10 items-center rounded-xl border border-red-200 px-3 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-60"
        >
          Obriši
        </button>
      </div>
    </article>
  );
}
