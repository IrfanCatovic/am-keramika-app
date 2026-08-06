"use client";

import { Category } from "@/types/category";
import { ProductGroup } from "@/types/product-group";

export type ProductFilterValues = {
  search: string;
  categoryID: number | null;
  groupID: number | null;
  ungrouped: boolean;
  includeInactive: boolean;
};

export function ProductFilters({
  values,
  categories,
  groups,
  groupsLoading,
  onChange,
  onSearchChange,
}: {
  values: ProductFilterValues;
  categories: Category[];
  groups: ProductGroup[];
  groupsLoading: boolean;
  onChange: (patch: Partial<ProductFilterValues>) => void;
  onSearchChange: (search: string) => void;
}) {
  return (
    <div className="rounded-2xl border border-stone-200/90 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)] sm:p-5">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <div className="sm:col-span-2 lg:col-span-2">
          <label
            htmlFor="product-search"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Pretraga
          </label>
          <input
            id="product-search"
            type="search"
            value={values.search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Naziv proizvoda..."
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2"
          />
        </div>

        <div>
          <label
            htmlFor="product-category-filter"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Kategorija
          </label>
          <select
            id="product-category-filter"
            value={values.categoryID ?? ""}
            onChange={(event) => {
              const next = event.target.value
                ? Number(event.target.value)
                : null;
              onChange({
                categoryID: next && Number.isFinite(next) ? next : null,
                groupID: null,
              });
            }}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2"
          >
            <option value="">Sve kategorije</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
                {!category.isActive ? " (neaktivna)" : ""}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label
            htmlFor="product-group-filter"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Grupa
          </label>
          <select
            id="product-group-filter"
            value={values.groupID ?? ""}
            disabled={!values.categoryID || values.ungrouped || groupsLoading}
            onChange={(event) => {
              const next = event.target.value
                ? Number(event.target.value)
                : null;
              onChange({
                groupID: next && Number.isFinite(next) ? next : null,
                ungrouped: false,
              });
            }}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
          >
            <option value="">
              {!values.categoryID
                ? "Izaberite kategoriju"
                : groupsLoading
                  ? "Učitavanje..."
                  : "Sve grupe"}
            </option>
            {groups.map((group) => (
              <option key={group.id} value={group.id}>
                {group.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
        <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 text-sm text-stone-700">
          <input
            type="checkbox"
            checked={values.ungrouped}
            onChange={(event) =>
              onChange({
                ungrouped: event.target.checked,
                groupID: event.target.checked ? null : values.groupID,
              })
            }
            className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
          />
          Samo bez grupe
        </label>
        <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 text-sm text-stone-700">
          <input
            type="checkbox"
            checked={values.includeInactive}
            onChange={(event) =>
              onChange({ includeInactive: event.target.checked })
            }
            className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
          />
          Prikaži neaktivne
        </label>
      </div>
    </div>
  );
}
