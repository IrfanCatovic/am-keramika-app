import { formatMoney } from "@/lib/format";
import { resolveProductUnitPrice } from "@/lib/product-pricing";

type ProductPriceFields = {
  salePrice: number;
  effectiveSalePrice?: number;
  isOnSale?: boolean;
  discountPercent?: number;
};

export function ProductSalePrice({
  product,
  className = "",
}: {
  product: ProductPriceFields;
  className?: string;
}) {
  const effective = resolveProductUnitPrice(product);
  const onSale = Boolean(product.isOnSale) && (product.discountPercent ?? 0) > 0;

  if (!onSale) {
    return (
      <span className={`tabular-nums font-medium text-stone-900 ${className}`}>
        {formatMoney(effective)}
      </span>
    );
  }

  const discount = Math.round(product.discountPercent ?? 0);

  return (
    <span className={`inline-flex flex-wrap items-baseline gap-x-2 gap-y-0.5 ${className}`}>
      {effective !== product.salePrice ? (
        <span className="tabular-nums text-stone-400 line-through">
          {formatMoney(product.salePrice)}
        </span>
      ) : null}
      <span className="tabular-nums font-semibold text-stone-900">
        {formatMoney(effective)}
      </span>
      <span className="rounded bg-rose-50 px-1.5 py-0.5 text-xs font-medium text-rose-700">
        -{discount}%
      </span>
    </span>
  );
}
