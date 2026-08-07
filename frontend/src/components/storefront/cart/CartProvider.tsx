"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import {
  clearCartStorage,
  readCartFromStorage,
  writeCartToStorage,
} from "@/lib/cart-storage";
import type { CartAddInput, CartItem } from "@/types/cart";

type CartContextValue = {
  items: CartItem[];
  hydrated: boolean;
  drawerOpen: boolean;
  /** Broj različitih proizvoda (badge). */
  itemCount: number;
  feedback: string | null;
  openDrawer: () => void;
  closeDrawer: () => void;
  toggleDrawer: () => void;
  clearFeedback: () => void;
  addItem: (input: CartAddInput) => void;
  setItemQuantity: (productId: number, quantity: number) => void;
  removeItem: (productId: number) => void;
  clearCart: () => void;
  updateItemSnapshot: (productId: number, patch: Partial<CartItem>) => void;
  getItem: (productId: number) => CartItem | undefined;
};

const CartContext = createContext<CartContextValue | null>(null);

function roundQty(q: number): number {
  return Math.round(q * 100) / 100;
}

export function CartProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<CartItem[]>([]);
  const [hydrated, setHydrated] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    const frame = window.requestAnimationFrame(() => {
      if (cancelled) return;
      setItems(readCartFromStorage());
      setHydrated(true);
    });
    return () => {
      cancelled = true;
      window.cancelAnimationFrame(frame);
    };
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    writeCartToStorage(items);
  }, [items, hydrated]);

  useEffect(() => {
    if (!feedback) return;
    const timer = window.setTimeout(() => setFeedback(null), 4200);
    return () => window.clearTimeout(timer);
  }, [feedback]);

  const openDrawer = useCallback(() => setDrawerOpen(true), []);
  const closeDrawer = useCallback(() => setDrawerOpen(false), []);
  const toggleDrawer = useCallback(() => setDrawerOpen((v) => !v), []);
  const clearFeedback = useCallback(() => setFeedback(null), []);

  const addItem = useCallback((input: CartAddInput) => {
    const qty = roundQty(input.quantity);
    if (!(qty > 0)) return;

    setItems((prev) => {
      const index = prev.findIndex((item) => item.productId === input.productId);
      if (index === -1) {
        return [
          ...prev,
          {
            productId: input.productId,
            slug: input.slug,
            name: input.name,
            imageUrl: input.imageUrl,
            unit: input.unit,
            quantity: qty,
            salePrice: input.salePrice,
            effectiveSalePrice: input.effectiveSalePrice,
            isOnSale: input.isOnSale,
            discountPercent: input.discountPercent,
            categoryName: input.categoryName,
            groupName: input.groupName,
          },
        ];
      }
      const next = [...prev];
      const existing = next[index];
      next[index] = {
        ...existing,
        ...input,
        quantity: roundQty(existing.quantity + qty),
      };
      return next;
    });
    setFeedback("Proizvod je dodat u korpu.");
    setDrawerOpen(true);
  }, []);

  const setItemQuantity = useCallback((productId: number, quantity: number) => {
    const qty = roundQty(quantity);
    setItems((prev) => {
      if (!(qty > 0)) {
        return prev.filter((item) => item.productId !== productId);
      }
      return prev.map((item) =>
        item.productId === productId ? { ...item, quantity: qty } : item,
      );
    });
  }, []);

  const removeItem = useCallback((productId: number) => {
    setItems((prev) => prev.filter((item) => item.productId !== productId));
  }, []);

  const clearCart = useCallback(() => {
    setItems([]);
    clearCartStorage();
  }, []);

  const updateItemSnapshot = useCallback(
    (productId: number, patch: Partial<CartItem>) => {
      setItems((prev) =>
        prev.map((item) =>
          item.productId === productId ? { ...item, ...patch } : item,
        ),
      );
    },
    [],
  );

  const getItem = useCallback(
    (productId: number) => items.find((item) => item.productId === productId),
    [items],
  );

  const value = useMemo<CartContextValue>(
    () => ({
      items,
      hydrated,
      drawerOpen,
      itemCount: items.length,
      feedback,
      openDrawer,
      closeDrawer,
      toggleDrawer,
      clearFeedback,
      addItem,
      setItemQuantity,
      removeItem,
      clearCart,
      updateItemSnapshot,
      getItem,
    }),
    [
      items,
      hydrated,
      drawerOpen,
      feedback,
      openDrawer,
      closeDrawer,
      toggleDrawer,
      clearFeedback,
      addItem,
      setItemQuantity,
      removeItem,
      clearCart,
      updateItemSnapshot,
      getItem,
    ],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext);
  if (!ctx) {
    throw new Error("useCart must be used within CartProvider");
  }
  return ctx;
}
