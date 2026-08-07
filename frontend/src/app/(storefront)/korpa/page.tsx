import type { Metadata } from "next";

import { CartPageClient } from "@/components/storefront/cart/CartPageClient";

export const metadata: Metadata = {
  title: "Korpa",
  description: "Pregledajte proizvode u vašoj korpi — AM Keramika.",
  robots: { index: false, follow: false },
};

export default function CartPage() {
  return <CartPageClient />;
}
