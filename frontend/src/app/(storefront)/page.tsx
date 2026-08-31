import {
  CategoryShowcase,
  FinalCtaSection,
  ProductSection,
  SalonSection,
  StorefrontHero,
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
    safeFetchPublicProducts({ homepage: true, limit: 8 }, cacheOptions),
    safeFetchPublicProducts({ onSale: true, limit: 8 }, cacheOptions),
    safeFetchPublicProducts({ random: true, limit: 8 }, cacheOptions),
  ]);

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
        title="Istaknuti proizvodi"
        products={featured?.products ?? []}
        href="/proizvodi"
        tone="default"
      />
      <SalonSection />
      <ProductSection
        eyebrow="Povoljno"
        title="Na akciji"
        products={onSale?.products ?? []}
        href="/proizvodi?onSale=true"
        tone="muted"
      />
      <ProductSection
        eyebrow="Inspiracija"
        title="Izdvajamo za vas"
        products={picks?.products ?? []}
        href="/proizvodi"
        tone="default"
      />
      <FinalCtaSection />
    </>
  );
}
