"use client";

import { useState } from "react";

import { useCustomerSearch } from "@/hooks/useCustomerSearch";
import { CustomerListItem } from "@/types/customer";

/**
 * Ponovo iskoristiv selector aktivnih kupaca (invoice forma kasnije).
 * Opcija „Gotovinska prodaja“ nije dio ove komponente.
 */
export function CustomerSelector({
  value,
  onChange,
  disabled = false,
  label = "Kupac",
}: {
  value: CustomerListItem | null;
  onChange: (customer: CustomerListItem | null) => void;
  disabled?: boolean;
  label?: string;
}) {
  const [query, setQuery] = useState(value?.name ?? "");
  const [open, setOpen] = useState(false);
  const displayQuery = value ? value.name : query;
  const { results, loading, error } = useCustomerSearch(
    value ? value.name : query,
    open || !value,
  );

  return (
    <div className="relative min-w-0">
      <label className="mb-1.5 block text-sm font-medium text-stone-700">
        {label}
      </label>
      <input
        value={displayQuery}
        disabled={disabled}
        onFocus={() => setOpen(true)}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
          if (value) {
            onChange(null);
          }
        }}
        placeholder="Pretraži po imenu ili telefonu"
        className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
      />
      {value ? (
        <p className="mt-1 text-xs text-stone-500">
          Izabrano: {value.name}
          {value.phone ? ` · ${value.phone}` : ""}
        </p>
      ) : null}
      {error ? <p className="mt-1 text-xs text-red-600">{error}</p> : null}
      {open && !disabled ? (
        <ul className="absolute z-20 mt-1 max-h-56 w-full overflow-auto rounded-xl border border-stone-200 bg-white py-1 shadow-lg">
          {loading ? (
            <li className="px-3 py-2 text-sm text-stone-500">Pretraga...</li>
          ) : null}
          {!loading && results.length === 0 ? (
            <li className="px-3 py-2 text-sm text-stone-500">Nema rezultata.</li>
          ) : null}
          {results.map((customer) => (
            <li key={customer.id}>
              <button
                type="button"
                className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-stone-50"
                onClick={() => {
                  onChange(customer);
                  setQuery(customer.name);
                  setOpen(false);
                }}
              >
                <span className="font-medium text-stone-900">
                  {customer.name}
                </span>
                {customer.phone ? (
                  <span className="text-xs text-stone-500">{customer.phone}</span>
                ) : null}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
