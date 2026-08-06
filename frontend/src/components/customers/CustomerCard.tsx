"use client";

import Link from "next/link";

import { CustomerListItem } from "@/types/customer";

export function CustomerCard({
  customer,
  includeInactive,
  busy,
  onActivate,
  onDeactivate,
  onDelete,
}: {
  customer: CustomerListItem;
  includeInactive: boolean;
  busy: boolean;
  onActivate: () => void;
  onDeactivate: () => void;
  onDelete: () => void;
}) {
  return (
    <article className="dash-enter min-w-0 rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="min-w-0">
        <Link
          href={`/customers/${customer.id}`}
          className="break-words text-base font-semibold text-stone-900 hover:text-[#8a6a45]"
        >
          {customer.name}
        </Link>
        <p className="mt-1 break-words text-sm text-stone-600">
          {customer.phone?.trim() ? customer.phone : "Bez telefona"}
        </p>
        {includeInactive ? (
          <p className="mt-2 text-xs text-stone-400">
            Pregled uključuje i neaktivne kupce.
          </p>
        ) : null}
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
        {includeInactive ? (
          <button
            type="button"
            disabled={busy}
            onClick={onActivate}
            className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
          >
            Aktiviraj
          </button>
        ) : null}
        <button
          type="button"
          disabled={busy}
          onClick={onDeactivate}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
        >
          Deaktiviraj
        </button>
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
