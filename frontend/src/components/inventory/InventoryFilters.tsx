"use client";

import { Category } from "@/types/category";
import { InventoryStockStatus } from "@/types/inventory";
import { ProductGroup } from "@/types/product-group";

export type InventoryFilterValues = {
  search: string;
  categoryID: number | null;
  groupID: number | null;
  status: InventoryStockStatus;
};

export function InventoryFilters({
  values,
  categories,
  groups,
  groupsLoading,
  onChange,
  onSearchChange,
  onReset,
}: {
  values: InventoryFilterValues;
  categories: Category[];
  groups: ProductGroup[];
  groupsLoading: boolean;
  onChange: (patch: Partial<InventoryFilterValues>) => void;
  onSearchChange: (search: string) => void;
  onReset: () => void;
}) {
  return (
    <div className="rounded-2xl border border-stone-200/90 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)] sm:p-5">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
        <div className="sm:col-span-2 lg:col-span-2">
          <label
            htmlFor="inventory-search"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Pretraga
          </label>
          <input
            id="inventory-search"
            type="search"
            value={values.search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Naziv proizvoda..."
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2"
          />
        </div>

        <div>
          <label
            htmlFor="inventory-category"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Kategorija
          </label>
          <select
            id="inventory-category"
            value={values.categoryID ?? ""}
            onChange={(event) => {
              const next = event.target.value
                ? Number(event.target.value)
                : null;
              onChange({
                categoryID: next,
                groupID: null,
              });
            }}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 focus:ring-2"
          >
            <option value="">Sve kategorije</option>
            {categories.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label
            htmlFor="inventory-group"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Grupa
          </label>
          <select
            id="inventory-group"
            value={values.groupID ?? ""}
            disabled={!values.categoryID || groupsLoading}
            onChange={(event) => {
              const next = event.target.value
                ? Number(event.target.value)
                : null;
              onChange({ groupID: next });
            }}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 focus:ring-2 disabled:bg-stone-50 disabled:text-stone-400"
          >
            <option value="">Sve grupe</option>
            {groups.map((group) => (
              <option key={group.id} value={group.id}>
                {group.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label
            htmlFor="inventory-status"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Status lagera
          </label>
          <select
            id="inventory-status"
            value={values.status}
            onChange={(event) =>
              onChange({
                status: event.target.value as InventoryStockStatus,
              })
            }
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 focus:ring-2"
          >
            <option value="all">Svi</option>
            <option value="low">Nizak lager</option>
            <option value="out">Nema na stanju</option>
          </select>
        </div>
      </div>

      <div className="mt-3 flex justify-end">
        <button
          type="button"
          onClick={onReset}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Reset filtera
        </button>
      </div>
    </div>
  );
}
