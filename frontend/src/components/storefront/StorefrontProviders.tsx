"use client";

import { CartDrawer } from "@/components/storefront/cart/CartDrawer";
import { CartProvider } from "@/components/storefront/cart/CartProvider";

export function StorefrontProviders({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <CartProvider>
      {children}
      <CartDrawer />
    </CartProvider>
  );
}
