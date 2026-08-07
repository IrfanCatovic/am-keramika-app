import type { Metadata } from "next";

import { CheckoutPageClient } from "@/components/storefront/checkout/CheckoutPageClient";

export const metadata: Metadata = {
  title: "Narudžbina",
  description: "Unesite podatke za online narudžbinu — AM Keramika.",
  robots: { index: false, follow: false },
};

export default function CheckoutPage() {
  return <CheckoutPageClient />;
}
