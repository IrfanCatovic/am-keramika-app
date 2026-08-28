import { Fraunces, Manrope } from "next/font/google";

import { StorefrontFooter } from "@/components/storefront/StorefrontFooter";
import { StorefrontHeader } from "@/components/storefront/StorefrontHeader";
import { StorefrontProviders } from "@/components/storefront/StorefrontProviders";
import { companyConfig } from "@/config/company";
import {
  PUBLIC_CATALOG_REVALIDATE_SECONDS,
  safeFetchPublicCategories,
} from "@/lib/public-catalog-api";

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

export default async function StorefrontLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const categories = await safeFetchPublicCategories({
    revalidate: PUBLIC_CATALOG_REVALIDATE_SECONDS,
  });

  return (
    <div
      className={`${display.variable} ${sans.variable} flex min-h-screen flex-col bg-[#f6f4f1] font-[family-name:var(--font-storefront-sans)] text-stone-900 antialiased`}
    >
      <StorefrontProviders>
        <StorefrontHeader categories={categories} />
        <main className="flex-1">{children}</main>
        <StorefrontFooter categories={categories} />
      </StorefrontProviders>
    </div>
  );
}
