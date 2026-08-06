"use client";

import {
  InvoiceSort,
  InvoiceSortDirection,
  InvoiceStatus,
} from "@/types/invoice";

export type InvoiceStatusFilter = InvoiceStatus | "";

export function InvoiceFilters({
  search,
  status,
  fromDate,
  toDate,
  sort,
  direction,
  onSearchChange,
  onStatusChange,
  onFromDateChange,
  onToDateChange,
  onSortChange,
  onDirectionChange,
  onReset,
}: {
  search: string;
  status: InvoiceStatusFilter;
  fromDate: string;
  toDate: string;
  sort: InvoiceSort;
  direction: InvoiceSortDirection;
  onSearchChange: (value: string) => void;
  onStatusChange: (value: InvoiceStatusFilter) => void;
  onFromDateChange: (value: string) => void;
  onToDateChange: (value: string) => void;
  onSortChange: (value: InvoiceSort) => void;
  onDirectionChange: (value: InvoiceSortDirection) => void;
  onReset: () => void;
}) {
  return (
    <section className="dash-enter rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        <div className="sm:col-span-2 xl:col-span-2">
          <label className="mb-1.5 block text-sm font-medium text-stone-700">
            Pretraga
          </label>
          <input
            value={search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Ime kupca ili broj računa"
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
          />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-stone-700">
            Status
          </label>
          <select
            value={status}
            onChange={(event) =>
              onStatusChange(event.target.value as InvoiceStatusFilter)
            }
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm"
          >
            <option value="">Svi statusi</option>
            <option value="unpaid">Neplaćeno</option>
            <option value="partially_paid">Djelimično</option>
            <option value="paid">Plaćeno</option>
            <option value="cancelled">Stornirano</option>
          </select>
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-stone-700">
            Od datuma
          </label>
          <input
            type="date"
            value={fromDate}
            onChange={(event) => onFromDateChange(event.target.value)}
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm"
          />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-stone-700">
            Do datuma
          </label>
          <input
            type="date"
            value={toDate}
            onChange={(event) => onToDateChange(event.target.value)}
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm"
          />
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-stone-700">
            Sortiranje
          </label>
          <div className="flex gap-2">
            <select
              value={sort}
              onChange={(event) =>
                onSortChange(event.target.value as InvoiceSort)
              }
              className="min-w-0 flex-1 rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm"
            >
              <option value="createdAt">Datum</option>
              <option value="totalAmount">Iznos</option>
            </select>
            <select
              value={direction}
              onChange={(event) =>
                onDirectionChange(event.target.value as InvoiceSortDirection)
              }
              className="w-[5.5rem] rounded-xl border border-stone-200 bg-white px-2 py-2.5 text-sm"
            >
              <option value="desc">Desc</option>
              <option value="asc">Asc</option>
            </select>
          </div>
        </div>
      </div>
      <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs text-stone-500">
          Pretraga: broj računa (ako je broj) ili ime kupca.
        </p>
        <button
          type="button"
          onClick={onReset}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Reset filtera
        </button>
      </div>
    </section>
  );
}
