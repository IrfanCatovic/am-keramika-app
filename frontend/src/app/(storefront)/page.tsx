import {
  CategoryShowcase,
  FinalCtaSection,
  ProductSection,
  SalonSection,
  StorefrontHero,
} from "@/components/storefront/StorefrontSections";
import {
  safeFetchPublicCategories,
  safeFetchPublicProducts,
} from "@/lib/public-catalog-api";

export const dynamic = "force-dynamic";

export default async function StorefrontHomePage() {
  const [categories, featured, onSale, picks] = await Promise.all([
    safeFetchPublicCategories(),
    safeFetchPublicProducts({ homepage: true, limit: 8 }),
    safeFetchPublicProducts({ onSale: true, limit: 8 }),
    safeFetchPublicProducts({ random: true, limit: 8 }),
  ]);

  return (
    <>
      <StorefrontHero />
      <CategoryShowcase categories={categories} />
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
