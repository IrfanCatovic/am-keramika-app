"use client";

import { formatMoney, formatQuantity } from "@/lib/format";
import { Product } from "@/types/product";

function productImageUrl(product: Product): string | null {
  return (
    product.primaryImage?.url ??
    product.images?.find((img) => img.isPrimary)?.url ??
    product.images?.[0]?.url ??
    null
  );
}

export function PosProductResultRow({
  product,
  selected,
  onSelect,
  disabledReason,
}: {
  product: Product;
  selected?: boolean;
  onSelect: () => void;
  disabledReason?: string | null;
}) {
  const imageUrl = productImageUrl(product);
  const outOfStock = product.stockQuantity <= 0;
  const disabled = Boolean(disabledReason) || outOfStock;
  const reason = disabledReason ?? (outOfStock ? "Nema na stanju" : null);

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onSelect}
      className={`flex w-full min-w-0 items-center gap-3 border-b border-stone-100 px-3 py-2.5 text-left transition last:border-b-0 ${
        selected && !disabled
          ? "bg-[#f3ebe1]"
          : disabled
            ? "cursor-not-allowed bg-stone-50/80 opacity-70"
            : "bg-white hover:bg-stone-50"
      }`}
    >
      <div className="h-11 w-11 shrink-0 overflow-hidden rounded-lg bg-stone-100 ring-1 ring-stone-200/80">
        {imageUrl ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={imageUrl} alt="" className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full w-full items-center justify-center text-[9px] text-stone-400">
            N/A
          </div>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-stone-900">
          {product.name}
        </p>
        <p className="mt-0.5 truncate text-xs text-stone-500">
          {product.category?.name ?? "—"}
          {product.group?.name ? ` · ${product.group.name}` : ""}
          {` · ${product.unit}`}
        </p>
      </div>
      <div className="shrink-0 text-right">
        <p className="text-sm font-semibold tabular-nums text-stone-900">
          {formatMoney(product.salePrice)}
        </p>
        <p
          className={`mt-0.5 text-[11px] tabular-nums ${
            outOfStock ? "font-medium text-red-700" : "text-stone-500"
          }`}
        >
          {outOfStock
            ? "Nema na stanju"
            : `Stanje ${formatQuantity(product.stockQuantity)}`}
        </p>
        {reason && !outOfStock ? (
          <p className="mt-0.5 text-[11px] text-amber-800">{reason}</p>
        ) : null}
      </div>
    </button>
  );
}

export function PosProductResults({
  products,
  activeIndex,
  onSelect,
  selectedQtyByProduct,
  emptyLabel = "Nema rezultata.",
}: {
  products: Product[];
  activeIndex: number;
  onSelect: (product: Product) => void;
  selectedQtyByProduct: Map<number, number>;
  emptyLabel?: string;
}) {
  if (products.length === 0) {
    return (
      <p className="px-3 py-4 text-sm text-stone-500">{emptyLabel}</p>
    );
  }

  return (
    <ul className="max-h-72 overflow-y-auto overscroll-contain">
      {products.map((product, index) => {
        const used = selectedQtyByProduct.get(product.id) ?? 0;
        const remaining = product.stockQuantity - used;
        const disabledReason =
          remaining <= 0 && product.stockQuantity > 0
            ? "Već na računu do maksimuma lagera"
            : null;
        return (
          <li key={product.id}>
            <PosProductResultRow
              product={product}
              selected={index === activeIndex}
              disabledReason={disabledReason}
              onSelect={() => onSelect(product)}
            />
          </li>
        );
      })}
    </ul>
  );
}
