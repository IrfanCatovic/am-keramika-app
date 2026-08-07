import {
  CategoryShowcase,
  ProductSection,
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
      />
      <ProductSection
        eyebrow="Povoljno"
        title="Na akciji"
        products={onSale?.products ?? []}
        href="/proizvodi?onSale=true"
      />
      <ProductSection
        eyebrow="Inspiracija"
        title="Izdvajamo za vas"
        products={picks?.products ?? []}
        href="/proizvodi"
      />
      <section className="border-t border-stone-200 bg-white/50 py-16">
        <div className="mx-auto max-w-3xl px-4 text-center sm:px-6">
          <h2 className="font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900">
            Kvalitet koji traje
          </h2>
          <p className="mt-4 text-sm leading-relaxed text-stone-500 sm:text-base">
            Pregledajte katalog i pronađite keramiku, sanitarije i opremu za
            svoj prostor. Online narudžba stiže uskoro — za sada istražite
            asortiman i detalje proizvoda.
          </p>
        </div>
      </section>
    </>
  );
}
