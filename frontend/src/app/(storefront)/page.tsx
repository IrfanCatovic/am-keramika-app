import {
  FinalCtaSection,
  ProductSection,
  SalonSection,
  StorefrontHero,
  TrustSection,
} from "@/components/storefront/StorefrontSections";
import {
  PUBLIC_CATALOG_REVALIDATE_SECONDS,
  safeFetchPublicProducts,
} from "@/lib/public-catalog-api";

export const revalidate = 60;

export default async function StorefrontHomePage() {
  const cacheOptions = {
    revalidate: PUBLIC_CATALOG_REVALIDATE_SECONDS,
  };
  const [featured, onSale, picks] = await Promise.all([
    safeFetchPublicProducts({ homepage: true, limit: 4 }, cacheOptions),
    safeFetchPublicProducts({ onSale: true, limit: 4 }, cacheOptions),
    safeFetchPublicProducts({ random: true, limit: 4 }, cacheOptions),
  ]);

  const featuredProducts = featured?.products ?? [];
  const saleProducts = onSale?.products ?? [];
  const reservedProductIds = new Set(
    [...featuredProducts, ...saleProducts].map((product) => product.id),
  );
  const additionalProducts = (picks?.products ?? []).filter(
    (product) => !reservedProductIds.has(product.id),
  );

  return (
    <>
      <StorefrontHero />
      <ProductSection
        eyebrow="Odabrano"
        title="Izdvojeno iz ponude"
        products={featuredProducts}
        href="/proizvodi"
        tone="default"
        homepage
      />
      <ProductSection
        eyebrow="Povoljno"
        title="Na akciji"
        products={saleProducts}
        href="/proizvodi?onSale=true"
        tone="muted"
        homepage
      />
      <SalonSection />
      <TrustSection />
      <ProductSection
        eyebrow="Inspiracija"
        title="Pogledajte još"
        products={additionalProducts}
        href="/proizvodi"
        tone="default"
        homepage
      />
      <FinalCtaSection />
    </>
  );
}
