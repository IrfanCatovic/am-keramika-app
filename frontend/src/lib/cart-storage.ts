import type { CartItem, CartPersistedState } from "@/types/cart";
import { getActualProductQuantity } from "@/lib/product-pricing";

export const CART_STORAGE_KEY = "am-keramika-cart-v2";
const LEGACY_CART_STORAGE_KEY = "am-keramika-cart-v1";

function isFinitePositive(n: unknown): n is number {
  return typeof n === "number" && Number.isFinite(n) && n > 0;
}

function normalizeItem(raw: unknown): CartItem | null {
  if (!raw || typeof raw !== "object") return null;
  const item = raw as Record<string, unknown>;
  const productId = Number(item.productId);
  const quantity = Number(item.quantity);
  const salePrice = Number(item.salePrice);
  const effectiveSalePrice = Number(item.effectiveSalePrice);
  const discountPercent = Number(item.discountPercent ?? 0);
  const saleByPackage = Boolean(item.saleByPackage);
  const packageQuantity = Number(item.packageQuantity ?? 0);
  const packagePrice =
    item.packagePrice == null ? null : Number(item.packagePrice);

  if (!Number.isInteger(productId) || productId <= 0) return null;
  if (!isFinitePositive(quantity)) return null;
  if (typeof item.slug !== "string" || !item.slug.trim()) return null;
  if (typeof item.name !== "string" || !item.name.trim()) return null;
  if (typeof item.unit !== "string") return null;
  if (!Number.isFinite(salePrice) || salePrice < 0) return null;
  if (!Number.isFinite(effectiveSalePrice) || effectiveSalePrice < 0) return null;
  if (!Number.isFinite(discountPercent) || discountPercent < 0) return null;
  if (
    !Number.isFinite(packageQuantity) ||
    packageQuantity < 0 ||
    (saleByPackage && packageQuantity <= 0)
  ) {
    return null;
  }
  if (packagePrice != null && (!Number.isFinite(packagePrice) || packagePrice < 0)) {
    return null;
  }

  const imageUrl =
    item.imageUrl === null
      ? null
      : typeof item.imageUrl === "string"
        ? item.imageUrl
        : null;

  const requestedQuantity = Math.round(quantity * 10000) / 10000;
  const details = getActualProductQuantity(
    requestedQuantity,
    saleByPackage,
    packageQuantity,
  );

  return {
    productId,
    slug: item.slug.trim(),
    name: item.name.trim(),
    imageUrl,
    unit: item.unit,
    quantity: requestedQuantity,
    requestedQuantity,
    actualQuantity: details.actualQuantity,
    packageCount: details.packageCount,
    packageQuantity,
    saleByPackage,
    packagePrice,
    salePrice,
    effectiveSalePrice,
    isOnSale: Boolean(item.isOnSale),
    discountPercent,
    categoryName:
      typeof item.categoryName === "string" ? item.categoryName : undefined,
    groupName: typeof item.groupName === "string" ? item.groupName : undefined,
  };
}

export function readCartFromStorage(): CartItem[] {
  if (typeof window === "undefined") return [];
  try {
    const rawValues = [
      window.localStorage.getItem(CART_STORAGE_KEY),
      window.localStorage.getItem(LEGACY_CART_STORAGE_KEY),
    ].filter((value): value is string => value != null);

    for (const raw of rawValues) {
      try {
        const parsed = JSON.parse(raw) as CartPersistedState | CartItem[];
        const items = Array.isArray(parsed)
          ? parsed
          : Array.isArray(parsed?.items)
            ? parsed.items
            : null;
        if (!items) continue;

        const normalized = items
          .map(normalizeItem)
          .filter((item): item is CartItem => item != null);
        const byId = new Map<number, CartItem>();
        for (const item of normalized) {
          const existing = byId.get(item.productId);
          if (existing) {
            const requestedQuantity =
              Math.round(
                (existing.requestedQuantity + item.requestedQuantity) * 10000,
              ) / 10000;
            const details = getActualProductQuantity(
              requestedQuantity,
              item.saleByPackage,
              item.packageQuantity,
            );
            byId.set(item.productId, {
              ...item,
              quantity: requestedQuantity,
              requestedQuantity,
              actualQuantity: details.actualQuantity,
              packageCount: details.packageCount,
            });
          } else {
            byId.set(item.productId, item);
          }
        }
        return Array.from(byId.values());
      } catch {
        // Try the legacy v1 payload if the current payload is malformed.
      }
    }
    window.localStorage.removeItem(CART_STORAGE_KEY);
    window.localStorage.removeItem(LEGACY_CART_STORAGE_KEY);
    return [];
  } catch {
    try {
      window.localStorage.removeItem(CART_STORAGE_KEY);
    } catch {
      /* ignore */
    }
    return [];
  }
}

export function writeCartToStorage(items: CartItem[]): void {
  if (typeof window === "undefined") return;
  try {
    const payload: CartPersistedState = { version: 2, items };
    window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(payload));
    window.localStorage.removeItem(LEGACY_CART_STORAGE_KEY);
  } catch {
    /* quota / private mode — ignore */
  }
}

export function clearCartStorage(): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.removeItem(CART_STORAGE_KEY);
    window.localStorage.removeItem(LEGACY_CART_STORAGE_KEY);
  } catch {
    /* ignore */
  }
}
