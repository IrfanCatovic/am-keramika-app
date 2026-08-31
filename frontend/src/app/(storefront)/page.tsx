import {
  CategoryShowcase,
  FinalCtaSection,
  ProductSection,
  SalonSection,
  StorefrontHero,
  TrustSection,
} from "@/components/storefront/StorefrontSections";
import {
  PUBLIC_CATALOG_REVALIDATE_SECONDS,
  safeFetchPublicCategories,
  safeFetchPublicProducts,
} from "@/lib/public-catalog-api";
import type { CategoryShowcaseItem } from "@/components/storefront/StorefrontSections";

export const revalidate = 60;

export default async function StorefrontHomePage() {
  const cacheOptions = {
    revalidate: PUBLIC_CATALOG_REVALIDATE_SECONDS,
  };
  const [categories, featured, onSale, picks] = await Promise.all([
    safeFetchPublicCategories(cacheOptions),
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

  const categoryProducts = await Promise.all(
    categories.map((category) =>
      safeFetchPublicProducts(
        { categorySlug: category.slug, limit: 4 },
        cacheOptions,
      ),
    ),
  );
  const categoryShowcaseItems: CategoryShowcaseItem[] = categories.map(
    (category, index) => ({
      ...category,
      image:
        categoryProducts[index]?.products.find(
          (product) => product.primaryImage?.url,
        )?.primaryImage ?? null,
    }),
  );

  return (
    <>
      <StorefrontHero />
      <CategoryShowcase categories={categoryShowcaseItems} />
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
