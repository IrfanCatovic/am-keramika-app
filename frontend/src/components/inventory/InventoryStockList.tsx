"use client";

import { InventoryProductThumb } from "@/components/inventory/InventoryProductThumb";
import { formatQuantity } from "@/lib/format";
import {
  getStockStatus,
  stockStatusClassName,
  stockStatusLabel,
} from "@/lib/inventory-status";
import { InventoryProductRow } from "@/types/inventory";

function categoryGroupLabel(product: InventoryProductRow): string {
  const parts = [product.category?.name, product.group?.name].filter(Boolean);
  return parts.length > 0 ? parts.join(" / ") : "Bez kategorije";
}

export function InventoryStockList({
  products,
  onAdjust,
}: {
  products: InventoryProductRow[];
  onAdjust: (product: InventoryProductRow) => void;
}) {
  if (products.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-5 py-10 text-center text-sm text-stone-500">
        Nema proizvoda za odabrane filtere.
      </div>
    );
  }

  return (
    <>
      <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b border-stone-200 bg-stone-50/80 text-left text-xs uppercase tracking-[0.08em] text-stone-500">
              <th className="px-4 py-3 font-semibold">Proizvod</th>
              <th className="px-4 py-3 font-semibold">Kategorija</th>
              <th className="px-4 py-3 font-semibold">Stanje</th>
              <th className="px-4 py-3 font-semibold">Minimum</th>
              <th className="px-4 py-3 font-semibold">Status</th>
              <th className="px-4 py-3 font-semibold">Akcije</th>
            </tr>
          </thead>
          <tbody>
            {products.map((product) => {
              const status = getStockStatus(
                product.stockQuantity,
                product.minStockQuantity,
              );
              return (
                <tr
                  key={product.id}
                  className="border-b border-stone-100 last:border-b-0"
                >
                  <td className="px-4 py-3">
                    <div className="flex min-w-0 items-center gap-3">
                      <InventoryProductThumb product={product} />
                      <div className="min-w-0">
                        <p className="truncate font-medium text-stone-900">
                          {product.name}
                        </p>
                        <p className="text-xs text-stone-500">{product.unit}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-stone-600">
                    {categoryGroupLabel(product)}
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-base font-semibold tabular-nums text-stone-900">
                      {formatQuantity(Math.max(0, product.stockQuantity))}
                    </span>
                    <span className="ml-1 text-stone-500">{product.unit}</span>
                  </td>
                  <td className="px-4 py-3 tabular-nums text-stone-700">
                    {formatQuantity(product.minStockQuantity)} {product.unit}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex rounded-full border px-2.5 py-0.5 text-xs font-medium ${stockStatusClassName(status)}`}
                    >
                      {stockStatusLabel(status)}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <button
                      type="button"
                      onClick={() => onAdjust(product)}
                      className="inline-flex min-h-9 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
                    >
                      Koriguj stanje
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <ul className="space-y-3 lg:hidden">
        {products.map((product) => {
          const status = getStockStatus(
            product.stockQuantity,
            product.minStockQuantity,
          );
          return (
            <li
              key={product.id}
              className="rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)]"
            >
              <div className="flex gap-3">
                <InventoryProductThumb product={product} />
                <div className="min-w-0 flex-1">
                  <p className="break-words font-medium text-stone-900">
                    {product.name}
                  </p>
                  <p className="mt-1 text-xs text-stone-500">
                    {categoryGroupLabel(product)}
                  </p>
                </div>
              </div>

              <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                <div>
                  <p className="text-xs text-stone-500">Stanje</p>
                  <p className="mt-1 text-lg font-semibold tabular-nums text-stone-900">
                    {formatQuantity(Math.max(0, product.stockQuantity))}{" "}
                    {product.unit}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-stone-500">Minimum</p>
                  <p className="mt-1 font-medium tabular-nums text-stone-800">
                    {formatQuantity(product.minStockQuantity)} {product.unit}
                  </p>
                </div>
              </div>

              <div className="mt-4 flex flex-wrap items-center justify-between gap-2">
                <span
                  className={`inline-flex rounded-full border px-2.5 py-0.5 text-xs font-medium ${stockStatusClassName(status)}`}
                >
                  {stockStatusLabel(status)}
                </span>
                <button
                  type="button"
                  onClick={() => onAdjust(product)}
                  className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50"
                >
                  Koriguj stanje
                </button>
              </div>
            </li>
          );
        })}
      </ul>
    </>
  );
}
