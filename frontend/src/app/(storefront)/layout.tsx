import { Fraunces, Manrope } from "next/font/google";

import { StorefrontFooter } from "@/components/storefront/StorefrontFooter";
import { StorefrontHeader } from "@/components/storefront/StorefrontHeader";
import { companyConfig } from "@/config/company";
import { safeFetchPublicCategories } from "@/lib/public-catalog-api";

import type { Metadata } from "next";

const display = Fraunces({
  subsets: ["latin"],
  variable: "--font-storefront-display",
  display: "swap",
});

const sans = Manrope({
  subsets: ["latin"],
  variable: "--font-storefront-sans",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: `${companyConfig.name} | Keramika, sanitarije i grijanje`,
    template: `%s | ${companyConfig.name}`,
  },
  description:
    "Keramika, sanitarije, grijanje i oprema za vaš dom. Pregledajte asortiman AM Keramika.",
};

export const dynamic = "force-dynamic";

export default async function StorefrontLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const categories = await safeFetchPublicCategories();

  return (
    <div
      className={`${display.variable} ${sans.variable} flex min-h-screen flex-col bg-[#f7f5f2] font-[family-name:var(--font-storefront-sans)] text-stone-900 antialiased`}
    >
      <StorefrontHeader categories={categories} />
      <main className="flex-1">{children}</main>
      <StorefrontFooter categories={categories} />
    </div>
  );
}
