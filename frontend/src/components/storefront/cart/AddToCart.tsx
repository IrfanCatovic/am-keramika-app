"use client";

import { useId, useState } from "react";

import { CartQuantityControl } from "@/components/storefront/cart/CartQuantityControl";
import { useCart } from "@/components/storefront/cart/CartProvider";
import { useAvailabilityCheck } from "@/hooks/useAvailabilityCheck";
import { formatMoney, formatQuantity, formatUnit } from "@/lib/format";
import { getActualProductQuantity } from "@/lib/product-pricing";
import type { PublicProduct } from "@/types/public-catalog";

type AddToCartProduct = Pick<
  PublicProduct,
  | "id"
  | "slug"
  | "name"
  | "unit"
  | "salePrice"
  | "effectiveSalePrice"
  | "isOnSale"
  | "discountPercent"
  | "inStock"
  | "primaryImage"
  | "category"
  | "group"
  | "saleByPackage"
  | "packageQuantity"
  | "packagePrice"
>;

export function AddToCart({ product }: { product: AddToCartProduct }) {
  const qtyId = useId();
  const { addItem, getItem } = useCart();
  const { checking, error, clearError, checkNow, checkDebounced } =
    useAvailabilityCheck();
  const [quantity, setQuantity] = useState(1);
  const [busy, setBusy] = useState(false);
  const quantityDetails = getActualProductQuantity(
    quantity,
    product.saleByPackage,
    product.packageQuantity ?? 0,
  );

  const disabled = !product.inStock;

  async function handleQuantityChange(next: number) {
    clearError();
    setQuantity(next);
    if (!product.inStock || !(next > 0)) return;
    const existing = getItem(product.id)?.quantity ?? 0;
    checkDebounced(product.id, existing + next);
  }

  async function handleAdd() {
    if (!product.inStock || busy) return;
    clearError();
    setBusy(true);
    try {
      const existing = getItem(product.id)?.quantity ?? 0;
      const total = Math.round((existing + quantity) * 100) / 100;
      const result = await checkNow(product.id, total);
      if (!result.available || result.stale) return;
      addItem({
        productId: product.id,
        slug: product.slug,
        name: product.name,
        imageUrl: product.primaryImage?.url ?? null,
        unit: product.unit,
        quantity,
        salePrice: product.salePrice,
        effectiveSalePrice: product.effectiveSalePrice,
        isOnSale: product.isOnSale,
        discountPercent: product.discountPercent,
        saleByPackage: product.saleByPackage ?? false,
        packageQuantity: product.packageQuantity ?? 0,
        packagePrice: product.packagePrice,
        categoryName: product.category?.name,
        groupName: product.group?.name,
      });
      setQuantity(1);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mt-10 rounded-xl border border-stone-200 bg-[#f6f4f1] px-5 py-6">
      <p className="text-[11px] uppercase tracking-[0.16em] text-[#8a6a45]">
        Narudžbina
      </p>
      <p className="mt-2 text-sm text-stone-600">
        Izaberite količinu i dodajte proizvod u korpu.
      </p>

      <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center">
        <label htmlFor={qtyId} className="text-sm font-medium text-stone-700">
          Potrebna količina
        </label>
        <CartQuantityControl
          id={qtyId}
          value={quantity}
          unit={formatUnit(product.unit)}
          disabled={disabled || busy}
          onChange={(next) => {
            void handleQuantityChange(next);
          }}
        />
        <button
          type="button"
          disabled={disabled || busy || checking || !(quantity > 0)}
          onClick={() => void handleAdd()}
          className="inline-flex min-h-11 flex-1 items-center justify-center rounded-full bg-[#141311] px-6 text-sm font-medium text-white transition hover:bg-[#2a2420] disabled:cursor-not-allowed disabled:opacity-45 sm:flex-none sm:px-8"
        >
          {busy || checking ? "Provera…" : "Dodaj u korpu"}
        </button>
      </div>

      {product.saleByPackage && product.packageQuantity ? (
        <div className="mt-4 space-y-1 text-sm text-stone-600">
          <p>
            Potrebno: {formatQuantity(quantity)} {formatUnit(product.unit)}
          </p>
          <p>
            {quantityDetails.packageCount} paketa ×{" "}
            {formatQuantity(product.packageQuantity)} {formatUnit(product.unit)}
          </p>
          <p>
            Za obračun: {formatQuantity(quantityDetails.actualQuantity)}{" "}
            {formatUnit(product.unit)}
          </p>
          <p className="font-medium text-stone-800">
            Ukupno:{" "}
            {formatMoney(
              product.effectiveSalePrice * quantityDetails.actualQuantity,
            )}
          </p>
        </div>
      ) : null}

      {disabled ? (
        <p className="mt-3 text-sm text-stone-500">
          Trenutno nije na stanju
        </p>
      ) : null}
      {error ? (
        <p className="mt-3 text-sm text-stone-600" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
