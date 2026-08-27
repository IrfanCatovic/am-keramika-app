import { formatMoney } from "@/lib/format";

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
  showUnit = false,
  hidePackageLine = false,
}: {
  product: {
    salePrice: number;
    effectiveSalePrice: number;
    isOnSale?: boolean;
    discountPercent?: number;
    unit?: string;
    saleByPackage?: boolean;
    packagePrice?: number | null;
    packageQuantity?: number;
  };
  size?: "sm" | "md" | "lg";
  showUnit?: boolean;
  hidePackageLine?: boolean;
}) {
  const discount = product.discountPercent ?? 0;
  const onSale = Boolean(product.isOnSale) && discount > 0;
  const priceClass =
    size === "lg"
      ? "text-2xl font-semibold tracking-tight"
      : size === "sm"
        ? "text-sm font-semibold"
        : "text-base font-semibold";
  const strikeClass =
    size === "lg" ? "text-sm" : size === "sm" ? "text-xs" : "text-sm";
  const unitSuffix =
    showUnit && product.unit ? ` / ${product.unit}` : "";

  return (
    <div>
      <div className="inline-flex flex-wrap items-baseline gap-x-2 gap-y-1">
        {onSale ? (
          <span className={`tabular-nums text-stone-400 line-through ${strikeClass}`}>
            {formatMoney(product.salePrice)}
            {unitSuffix}
          </span>
        ) : null}
        <span className={`tabular-nums text-stone-900 ${priceClass}`}>
          {formatMoney(product.effectiveSalePrice)}
          {unitSuffix}
        </span>
        {onSale ? (
          <span className="rounded-md bg-[#f1ebe4] px-1.5 py-0.5 text-xs font-medium text-[#5c4630]">
            -{Math.round(discount)}%
          </span>
        ) : null}
      </div>
      {!hidePackageLine && product.saleByPackage && product.packagePrice ? (
        <p className="mt-0.5 text-xs font-medium text-stone-500">
          {formatMoney(product.packagePrice)} / paket
        </p>
      ) : null}
    </div>
  );
}
