"use client";

import Link from "next/link";

import { CustomerCard } from "@/components/customers/CustomerCard";
import { CustomerStatusBadge } from "@/components/customers/CustomerStatusBadge";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { CustomerListItem } from "@/types/customer";

export function CustomerList({
  customers,
  loading,
  error,
  searchActive,
  busyId,
  onRetry,
  onActivate,
  onDeactivate,
  onDelete,
}: {
  customers: CustomerListItem[];
  loading: boolean;
  error: string | null;
  searchActive: boolean;
  busyId: number | null;
  onRetry: () => void;
  onActivate: (customer: CustomerListItem) => void;
  onDeactivate: (customer: CustomerListItem) => void;
  onDelete: (customer: CustomerListItem) => void;
}) {
  if (loading) {
    return <ListSkeleton rows={5} />;
  }

  if (error) {
    return <InlineError message={error} onRetry={onRetry} />;
  }

  if (customers.length === 0) {
    return (
      <EmptyState
        title={searchActive ? "Nema rezultata pretrage" : "Nema kupaca"}
        description={
          searchActive
            ? "Pokušajte drugačiji pojam ili promijenite status filter."
            : "Dodajte prvog kupca da biste evidentirali dugovanja i račune."
        }
        action={
          !searchActive ? (
            <Link
              href="/customers/new"
              className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
            >
              Novi kupac
            </Link>
          ) : undefined
        }
      />
    );
  }

  return (
    <>
      <ul className="space-y-3 md:hidden">
        {customers.map((customer) => (
          <li key={customer.id}>
            <CustomerCard
              customer={customer}
              busy={busyId === customer.id}
              onActivate={() => onActivate(customer)}
              onDeactivate={() => onDeactivate(customer)}
              onDelete={() => onDelete(customer)}
            />
          </li>
        ))}
      </ul>

      <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white md:block">
        <table className="w-full table-fixed text-left text-sm">
          <thead className="sticky top-0 bg-stone-50/95 backdrop-blur">
            <tr className="border-b border-stone-200 text-xs uppercase tracking-[0.08em] text-stone-500">
              <th className="w-[32%] px-4 py-3 font-medium">Naziv</th>
              <th className="w-[20%] px-4 py-3 font-medium">Telefon</th>
              <th className="w-[14%] px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Akcije</th>
            </tr>
          </thead>
          <tbody>
            {customers.map((customer) => {
              const muted = !customer.isActive;
              return (
                <tr
                  key={customer.id}
                  className={`border-b border-stone-100 last:border-b-0 ${
                    muted ? "bg-stone-50/80 text-stone-500" : ""
                  }`}
                >
                  <td className="px-4 py-3 align-top">
                    <Link
                      href={`/customers/${customer.id}`}
                      className={`break-words font-medium hover:text-[#8a6a45] ${
                        muted ? "text-stone-600" : "text-stone-900"
                      }`}
                    >
                      {customer.name}
                    </Link>
                  </td>
                  <td className="break-words px-4 py-3 align-top text-stone-600">
                    {customer.phone?.trim() ? customer.phone : "—"}
                  </td>
                  <td className="px-4 py-3 align-top">
                    <CustomerStatusBadge isActive={customer.isActive} />
                  </td>
                  <td className="px-4 py-3 align-top">
                    <div className="flex flex-wrap gap-2">
                      <Link
                        href={`/customers/${customer.id}`}
                        className="rounded-lg border border-stone-200 px-2.5 py-1.5 text-xs font-medium text-stone-700 hover:bg-stone-50"
                      >
                        Detalji
                      </Link>
                      <Link
                        href={`/customers/${customer.id}/edit`}
                        className="rounded-lg border border-stone-200 px-2.5 py-1.5 text-xs font-medium text-stone-700 hover:bg-stone-50"
                      >
                        Uredi
                      </Link>
                      {customer.isActive ? (
                        <button
                          type="button"
                          disabled={busyId === customer.id}
                          onClick={() => onDeactivate(customer)}
                          className="rounded-lg border border-stone-200 px-2.5 py-1.5 text-xs font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
                        >
                          Deaktiviraj
                        </button>
                      ) : (
                        <button
                          type="button"
                          disabled={busyId === customer.id}
                          onClick={() => onActivate(customer)}
                          className="rounded-lg border border-stone-200 px-2.5 py-1.5 text-xs font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
                        >
                          Aktiviraj
                        </button>
                      )}
                      <button
                        type="button"
                        disabled={busyId === customer.id}
                        onClick={() => onDelete(customer)}
                        className="rounded-lg border border-red-200 px-2.5 py-1.5 text-xs font-medium text-red-700 hover:bg-red-50 disabled:opacity-60"
                      >
                        Obriši
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </>
  );
}
