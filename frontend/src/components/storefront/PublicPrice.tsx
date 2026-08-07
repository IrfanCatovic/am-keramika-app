import { formatMoney } from "@/lib/format";
import type { PublicProduct } from "@/types/public-catalog";

export function PublicAvailability({
  inStock,
  className = "",
}: {
  inStock: boolean;
  className?: string;
}) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 text-xs ${
        inStock ? "text-stone-600" : "text-stone-400"
      } ${className}`}
    >
      <span
        className={`h-1.5 w-1.5 rounded-full ${
          inStock ? "bg-emerald-600" : "bg-stone-300"
        }`}
        aria-hidden
      />
      {inStock ? "Na stanju" : "Trenutno nije na stanju"}
    </span>
  );
}

export function PublicProductPrice({
  product,
  size = "md",
}: {
  product: Pick<
    PublicProduct,
    "salePrice" | "effectiveSalePrice" | "isOnSale" | "discountPercent"
  >;
  size?: "sm" | "md" | "lg";
}) {
  const onSale = product.isOnSale && product.discountPercent > 0;
  const priceClass =
    size === "lg"
      ? "text-2xl font-semibold tracking-tight"
      : size === "sm"
        ? "text-sm font-semibold"
        : "text-base font-semibold";
  const strikeClass =
    size === "lg" ? "text-sm" : size === "sm" ? "text-xs" : "text-sm";

  if (!onSale) {
    return (
      <span className={`tabular-nums text-stone-900 ${priceClass}`}>
        {formatMoney(product.effectiveSalePrice)}
      </span>
    );
  }

  return (
    <span className="inline-flex flex-wrap items-baseline gap-x-2 gap-y-1">
      <span className={`tabular-nums text-stone-400 line-through ${strikeClass}`}>
        {formatMoney(product.salePrice)}
      </span>
      <span className={`tabular-nums text-stone-900 ${priceClass}`}>
        {formatMoney(product.effectiveSalePrice)}
      </span>
      <span className="rounded-md bg-[#f3ebe3] px-1.5 py-0.5 text-xs font-medium text-[#6b4f32]">
        -{Math.round(product.discountPercent)}%
      </span>
    </span>
  );
}
