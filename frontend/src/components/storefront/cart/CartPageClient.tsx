"use client";

import { useEffect, useMemo, useRef, useState } from "react";

import { CartEmptyState } from "@/components/storefront/cart/CartEmptyState";
import { CartItemRow } from "@/components/storefront/cart/CartItem";
import { useCart } from "@/components/storefront/cart/CartProvider";
import { CartSummary } from "@/components/storefront/cart/CartSummary";
import { fetchPublicProductBySlug } from "@/lib/public-catalog-api";
import type { CartItem } from "@/types/cart";

type Meta = {
  unavailable: boolean;
  priceUpdated: boolean;
};

export function CartPageClient() {
  const { items, hydrated, clearCart, updateItemSnapshot } = useCart();
  const [metaById, setMetaById] = useState<Record<number, Meta>>({});
  const [refreshing, setRefreshing] = useState(false);
  const syncedKey = useRef("");

  const productKey = items.map((i) => i.productId).join(",");

  useEffect(() => {
    if (!hydrated) return;
    if (!productKey) {
      syncedKey.current = "";
      return;
    }
    if (syncedKey.current === productKey) return;
    syncedKey.current = productKey;

    let cancelled = false;

    async function refresh() {
      setRefreshing(true);
      const snapshot = items;
      const nextMeta: Record<number, Meta> = {};

      await Promise.all(
        snapshot.map(async (item) => {
          try {
            const product = await fetchPublicProductBySlug(item.slug);
            if (product.id !== item.productId) {
              nextMeta[item.productId] = {
                unavailable: true,
                priceUpdated: false,
              };
              return;
            }

            const priceUpdated =
              product.effectiveSalePrice !== item.effectiveSalePrice ||
              product.salePrice !== item.salePrice ||
              product.isOnSale !== item.isOnSale ||
              product.discountPercent !== item.discountPercent;

            const patch: Partial<CartItem> = {
              name: product.name,
              slug: product.slug,
              unit: product.unit,
              imageUrl: product.primaryImage?.url ?? item.imageUrl,
              salePrice: product.salePrice,
              effectiveSalePrice: product.effectiveSalePrice,
              isOnSale: product.isOnSale,
              discountPercent: product.discountPercent,
              categoryName: product.category?.name,
              groupName: product.group?.name,
            };

            updateItemSnapshot(item.productId, patch);
            nextMeta[item.productId] = {
              unavailable: false,
              priceUpdated,
            };
          } catch {
            nextMeta[item.productId] = {
              unavailable: true,
              priceUpdated: false,
            };
          }
        }),
      );

      if (!cancelled) {
        setMetaById(nextMeta);
        setRefreshing(false);
      }
    }

    void refresh();
    return () => {
      cancelled = true;
    };
    // Only re-sync when the set of product IDs changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hydrated, productKey]);

  const resolvedMeta = useMemo(
    () => (productKey ? metaById : {}),
    [productKey, metaById],
  );

  const availableItems = useMemo(
    () => items.filter((item) => !resolvedMeta[item.productId]?.unavailable),
    [items, resolvedMeta],
  );

  const subtotal = availableItems.reduce(
    (sum, item) => sum + item.effectiveSalePrice * item.quantity,
    0,
  );

  const savings = availableItems.reduce((sum, item) => {
    if (item.isOnSale && item.discountPercent > 0) {
      return (
        sum +
        Math.max(
          0,
          item.salePrice * item.quantity -
            item.effectiveSalePrice * item.quantity,
        )
      );
    }
    return sum;
  }, 0);

  if (!hydrated) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <div className="h-8 w-48 animate-pulse rounded bg-stone-200" />
        <div className="mt-8 h-40 animate-pulse rounded-xl bg-stone-100" />
      </div>
    );
  }

  if (items.length === 0) {
    return <CartEmptyState />;
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-[11px] uppercase tracking-[0.18em] text-[#8a6a45]">
            Korpa
          </p>
          <h1 className="mt-2 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900 sm:text-4xl">
            Vaša korpa
          </h1>
          {refreshing ? (
            <p className="mt-2 text-xs text-stone-400">Ažuriranje cena…</p>
          ) : null}
        </div>
        {items.length > 1 ? (
          <button
            type="button"
            className="text-sm text-stone-500 underline-offset-4 hover:text-stone-800 hover:underline"
            onClick={() => {
              if (
                window.confirm(
                  "Da li ste sigurni da želite da ispraznite korpu?",
                )
              ) {
                clearCart();
              }
            }}
          >
            Isprazni korpu
          </button>
        ) : null}
      </div>

      <div className="grid gap-10 lg:grid-cols-[minmax(0,1fr)_320px] xl:grid-cols-[minmax(0,1fr)_360px]">
        <div>
          {items.map((item) => (
            <CartItemRow
              key={item.productId}
              item={item}
              unavailable={resolvedMeta[item.productId]?.unavailable ?? false}
              priceUpdated={resolvedMeta[item.productId]?.priceUpdated ?? false}
            />
          ))}
        </div>
        <CartSummary
          productCount={availableItems.length}
          subtotal={subtotal}
          savings={savings}
        />
      </div>
    </div>
  );
}
