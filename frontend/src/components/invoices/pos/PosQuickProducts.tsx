"use client";

import { ProductSalePrice } from "@/components/products/ProductSalePrice";
import { formatQuantity } from "@/lib/format";
import { Product } from "@/types/product";

function productImageUrl(product: Product): string | null {
  return (
    product.primaryImage?.url ??
    product.images?.find((img) => img.isPrimary)?.url ??
    product.images?.[0]?.url ??
    null
  );
}

export function PosQuickProducts({
  products,
  loading,
  onSelect,
  selectedQtyByProduct,
  title = "Brzi izbor",
}: {
  products: Product[];
  loading?: boolean;
  onSelect: (product: Product) => void;
  selectedQtyByProduct: Map<number, number>;
  title?: string;
}) {
  return (
    <section className="min-w-0">
      <div className="mb-2 flex items-baseline justify-between gap-2">
        <h2 className="text-sm font-semibold text-stone-800">{title}</h2>
        {loading ? (
          <span className="text-xs text-stone-400">Učitavanje…</span>
        ) : null}
      </div>

      {products.length === 0 && !loading ? (
        <p className="rounded-2xl border border-dashed border-stone-200 bg-white px-4 py-8 text-center text-sm text-stone-500">
          Nema proizvoda za brzi izbor. Koristite pretragu iznad.
        </p>
      ) : (
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4">
          {products.map((product) => {
            const imageUrl = productImageUrl(product);
            const used = selectedQtyByProduct.get(product.id) ?? 0;
            const remaining = product.stockQuantity - used;
            const outOfStock = product.stockQuantity <= 0;
            const disabled = outOfStock || remaining <= 0;

            return (
              <button
                key={product.id}
                type="button"
                disabled={disabled}
                onClick={() => onSelect(product)}
                className={`min-w-0 rounded-xl border p-2.5 text-left transition ${
                  disabled
                    ? "cursor-not-allowed border-stone-100 bg-stone-50 opacity-60"
                    : "border-stone-200 bg-white hover:border-[#c4a484]/70 hover:bg-[#faf6f1] active:scale-[0.99]"
                }`}
              >
                <div className="mb-2 aspect-[4/3] overflow-hidden rounded-lg bg-stone-100">
                  {imageUrl ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={imageUrl}
                      alt=""
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center text-[10px] text-stone-400">
                      N/A
                    </div>
                  )}
                </div>
                <p className="line-clamp-2 text-xs font-medium text-stone-900">
                  {product.name}
                </p>
                <p className="mt-1 text-xs">
                  <ProductSalePrice product={product} />
                </p>
                <p
                  className={`mt-0.5 text-[10px] tabular-nums ${
                    outOfStock ? "text-red-700" : "text-stone-500"
                  }`}
                >
                  {outOfStock
                    ? "Nema na stanju"
                    : `Stanje ${formatQuantity(product.stockQuantity)}`}
                </p>
              </button>
            );
          })}
        </div>
      )}
    </section>
  );
}
