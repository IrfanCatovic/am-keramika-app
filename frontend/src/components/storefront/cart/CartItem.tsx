"use client";

import Link from "next/link";

import { CartQuantityControl } from "@/components/storefront/cart/CartQuantityControl";
import { useCart } from "@/components/storefront/cart/CartProvider";
import { PublicProductPrice } from "@/components/storefront/PublicPrice";
import { useAvailabilityCheck } from "@/hooks/useAvailabilityCheck";
import { formatMoney } from "@/lib/format";
import type { CartItem as CartItemType } from "@/types/cart";

export function CartItemRow({
  item,
  unavailable = false,
  priceUpdated = false,
}: {
  item: CartItemType;
  unavailable?: boolean;
  priceUpdated?: boolean;
}) {
  const { setItemQuantity, removeItem } = useCart();
  const { checking, error, clearError, checkNow } = useAvailabilityCheck();

  const meta = [item.categoryName, item.groupName].filter(Boolean).join(" · ");

  async function applyQuantity(next: number) {
    clearError();
    if (!(next > 0)) return;
    if (unavailable) {
      setItemQuantity(item.productId, next);
      return;
    }
    const result = await checkNow(item.productId, next);
    if (!result.available || result.stale) return;
    setItemQuantity(item.productId, next);
  }

  return (
    <article
      className={`grid gap-4 border-b border-stone-200/90 py-6 sm:grid-cols-[112px_minmax(0,1fr)] ${
        unavailable ? "opacity-80" : ""
      }`}
    >
      <Link
        href={`/proizvodi/${item.slug}`}
        className="relative aspect-square overflow-hidden rounded-xl border border-stone-200 bg-white"
      >
        {item.imageUrl ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={item.imageUrl}
            alt=""
            className="h-full w-full object-contain p-3"
          />
        ) : (
          <div className="flex h-full items-center justify-center font-[family-name:var(--font-storefront-display)] text-2xl text-stone-300">
            AM
          </div>
        )}
      </Link>

      <div className="min-w-0">
        {meta ? (
          <p className="text-[10px] uppercase tracking-[0.16em] text-stone-400">
            {meta}
          </p>
        ) : null}
        <Link
          href={`/proizvodi/${item.slug}`}
          className="mt-1 block font-medium text-stone-900 transition hover:text-[#5c4630]"
        >
          {item.name}
        </Link>
        {item.unit ? (
          <p className="mt-1 text-sm text-stone-500">Jedinica: {item.unit}</p>
        ) : null}

        <div className="mt-3">
          <PublicProductPrice
            product={{
              salePrice: item.salePrice,
              effectiveSalePrice: item.effectiveSalePrice,
              isOnSale: item.isOnSale,
              discountPercent: item.discountPercent,
            }}
            size="sm"
          />
        </div>

        {unavailable ? (
          <p className="mt-3 text-sm text-stone-600" role="status">
            Proizvod trenutno nije dostupan.
          </p>
        ) : null}
        {priceUpdated && !unavailable ? (
          <p className="mt-2 text-xs text-stone-500">
            Cena proizvoda je ažurirana.
          </p>
        ) : null}

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <CartQuantityControl
            value={item.quantity}
            unit={item.unit}
            disabled={unavailable || checking}
            onChange={(next) => {
              void applyQuantity(next);
            }}
          />
          <p className="text-sm tabular-nums text-stone-900">
            {formatMoney(item.effectiveSalePrice * item.quantity)}
          </p>
          <button
            type="button"
            onClick={() => removeItem(item.productId)}
            className="text-sm text-stone-500 underline-offset-4 transition hover:text-stone-800 hover:underline"
          >
            Ukloni
          </button>
        </div>

        {error && !unavailable ? (
          <p className="mt-2 text-sm text-stone-600" role="alert">
            {error}
          </p>
        ) : null}
      </div>
    </article>
  );
}
