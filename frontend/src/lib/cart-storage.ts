import type { CartItem, CartPersistedState } from "@/types/cart";

export const CART_STORAGE_KEY = "am-keramika-cart-v1";

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

  if (!Number.isInteger(productId) || productId <= 0) return null;
  if (!isFinitePositive(quantity)) return null;
  if (typeof item.slug !== "string" || !item.slug.trim()) return null;
  if (typeof item.name !== "string" || !item.name.trim()) return null;
  if (typeof item.unit !== "string") return null;
  if (!Number.isFinite(salePrice) || salePrice < 0) return null;
  if (!Number.isFinite(effectiveSalePrice) || effectiveSalePrice < 0) return null;
  if (!Number.isFinite(discountPercent) || discountPercent < 0) return null;

  const imageUrl =
    item.imageUrl === null
      ? null
      : typeof item.imageUrl === "string"
        ? item.imageUrl
        : null;

  return {
    productId,
    slug: item.slug.trim(),
    name: item.name.trim(),
    imageUrl,
    unit: item.unit,
    quantity: Math.round(quantity * 100) / 100,
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
    const raw = window.localStorage.getItem(CART_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as CartPersistedState | CartItem[];
    const items = Array.isArray(parsed)
      ? parsed
      : Array.isArray(parsed?.items)
        ? parsed.items
        : null;
    if (!items) {
      window.localStorage.removeItem(CART_STORAGE_KEY);
      return [];
    }
    const normalized = items
      .map(normalizeItem)
      .filter((item): item is CartItem => item != null);
    // Dedupe by productId — keep first, sum quantities carefully not needed; keep last wins
    const byId = new Map<number, CartItem>();
    for (const item of normalized) {
      const existing = byId.get(item.productId);
      if (existing) {
        byId.set(item.productId, {
          ...item,
          quantity: Math.round((existing.quantity + item.quantity) * 100) / 100,
        });
      } else {
        byId.set(item.productId, item);
      }
    }
    return Array.from(byId.values());
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
    const payload: CartPersistedState = { version: 1, items };
    window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(payload));
  } catch {
    /* quota / private mode — ignore */
  }
}

export function clearCartStorage(): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.removeItem(CART_STORAGE_KEY);
  } catch {
    /* ignore */
  }
}
