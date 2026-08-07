import Link from "next/link";
import { notFound } from "next/navigation";
import type { Metadata } from "next";

import { ProductGallery } from "@/components/storefront/ProductGallery";
import { AddToCart } from "@/components/storefront/cart/AddToCart";
import { PublicProductGrid } from "@/components/storefront/PublicProductCard";
import {
  PublicAvailability,
  PublicProductPrice,
} from "@/components/storefront/PublicPrice";
import { StorefrontBreadcrumb } from "@/components/storefront/StorefrontSections";
import {
  fetchPublicProductBySlug,
  safeFetchPublicProducts,
  PublicCatalogError,
} from "@/lib/public-catalog-api";

export const dynamic = "force-dynamic";

type Params = Promise<{ slug: string }>;

function truncate(text: string, max = 160): string {
  const cleaned = text.replace(/\s+/g, " ").trim();
  if (cleaned.length <= max) return cleaned;
  return `${cleaned.slice(0, max - 1).trimEnd()}…`;
}

export async function generateMetadata({
  params,
}: {
  params: Params;
}): Promise<Metadata> {
  const { slug } = await params;
  try {
    const product = await fetchPublicProductBySlug(slug);
    return {
      title: product.name,
      description: product.description
        ? truncate(product.description)
        : `${product.name} — AM Keramika asortiman.`,
    };
  } catch {
    return { title: "Proizvod" };
  }
}

export default async function ProductDetailPage({
  params,
}: {
  params: Params;
}) {
  const { slug } = await params;

  let product;
  try {
    product = await fetchPublicProductBySlug(slug);
  } catch (err) {
    if (err instanceof PublicCatalogError && err.status === 404) {
      notFound();
    }
    return (
      <div className="mx-auto max-w-3xl px-4 py-20 text-center">
        <h1 className="font-[family-name:var(--font-storefront-display)] text-3xl">
          Proizvod trenutno nije dostupan
        </h1>
        <p className="mt-3 text-sm text-stone-500">Pokušajte ponovo.</p>
        <Link
          href="/proizvodi"
          className="mt-6 inline-flex rounded-full bg-stone-900 px-5 py-2.5 text-sm text-white"
        >
          Nazad na proizvode
        </Link>
      </div>
    );
  }

  const images =
    product.images && product.images.length > 0
      ? product.images
      : product.primaryImage
        ? [product.primaryImage]
        : [];

  let related =
    (
      await safeFetchPublicProducts(
        product.group
          ? {
              groupSlug: product.group.slug,
              categorySlug: product.category?.slug,
              excludeId: product.id,
              limit: 8,
            }
          : product.category
            ? {
                categorySlug: product.category.slug,
                excludeId: product.id,
                limit: 8,
              }
            : { excludeId: product.id, limit: 8 },
      )
    )?.products ?? [];

  const relatedTitle = product.group
    ? `Još iz kolekcije ${product.group.name}`
    : product.category
      ? "Slični proizvodi"
      : null;

  if (!product.group && !product.category) {
    related = [];
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <StorefrontBreadcrumb
        items={[
          { label: "Početna", href: "/" },
          { label: "Proizvodi", href: "/proizvodi" },
          ...(product.category
            ? [
                {
                  label: product.category.name,
                  href: `/kategorije/${product.category.slug}`,
                },
              ]
            : []),
          { label: product.name },
        ]}
      />

      <div className="grid gap-10 lg:grid-cols-2 lg:gap-14">
        <ProductGallery images={images} productName={product.name} />

        <div>
          {(product.category || product.group) && (
            <p className="text-xs uppercase tracking-[0.16em] text-stone-400">
              {[product.category?.name, product.group?.name]
                .filter(Boolean)
                .join(" · ")}
            </p>
          )}
          <h1 className="mt-3 font-[family-name:var(--font-storefront-display)] text-4xl leading-tight text-stone-900">
            {product.name}
          </h1>

          <div className="mt-6 flex flex-wrap items-center gap-x-5 gap-y-2">
            <PublicProductPrice product={product} size="lg" />
            <PublicAvailability
              inStock={product.inStock}
              className="self-center"
            />
          </div>
          {product.unit ? (
            <p className="mt-3 text-sm text-stone-500">Jedinica: {product.unit}</p>
          ) : null}

          {product.description ? (
            <div className="mt-8 border-t border-stone-200 pt-8">
              <h2 className="text-sm font-medium text-stone-900">Opis</h2>
              <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-stone-600">
                {product.description}
              </p>
            </div>
          ) : null}

          {/* Cart actions */}
          <AddToCart product={product} />
        </div>
      </div>

      {relatedTitle && related.length > 0 ? (
        <section className="mt-20">
          <h2 className="mb-8 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900">
            {relatedTitle}
          </h2>
          <PublicProductGrid products={related} />
        </section>
      ) : null}
    </div>
  );
}
