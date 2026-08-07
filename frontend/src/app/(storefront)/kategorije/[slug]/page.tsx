import { Suspense } from "react";
import Link from "next/link";
import { notFound } from "next/navigation";
import type { Metadata } from "next";

import {
  CatalogFilters,
  CatalogPagination,
} from "@/components/storefront/CatalogFilters";
import { PublicProductGrid } from "@/components/storefront/PublicProductCard";
import {
  StorefrontBreadcrumb,
  StorefrontEmpty,
} from "@/components/storefront/StorefrontSections";
import {
  fetchPublicCategoryBySlug,
  fetchPublicProductGroups,
  safeFetchPublicProducts,
  PublicCatalogError,
} from "@/lib/public-catalog-api";
import type { PublicProductSort } from "@/types/public-catalog";

export const dynamic = "force-dynamic";

type Params = Promise<{ slug: string }>;
type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function first(value: string | string[] | undefined): string {
  if (Array.isArray(value)) return value[0] ?? "";
  return value ?? "";
}

export async function generateMetadata({
  params,
}: {
  params: Params;
}): Promise<Metadata> {
  const { slug } = await params;
  try {
    const category = await fetchPublicCategoryBySlug(slug);
    return {
      title: category.name,
      description: `Proizvodi iz kategorije ${category.name} — AM Keramika.`,
    };
  } catch {
    return { title: "Kategorija" };
  }
}

export default async function CategoryPage({
  params,
  searchParams,
}: {
  params: Params;
  searchParams: SearchParams;
}) {
  const { slug } = await params;
  const sp = await searchParams;
  const group = first(sp.group).trim();
  const search = first(sp.search).trim();
  const onSale = first(sp.onSale) === "true";
  const inStock = first(sp.inStock) === "true";
  const sort = (first(sp.sort) || "recommended") as PublicProductSort;
  const page = Math.max(1, Number(first(sp.page)) || 1);

  let category;
  try {
    category = await fetchPublicCategoryBySlug(slug);
  } catch (err) {
    if (err instanceof PublicCatalogError && err.status === 404) {
      notFound();
    }
    return (
      <div className="mx-auto max-w-3xl px-4 py-20 text-center">
        <h1 className="font-[family-name:var(--font-storefront-display)] text-3xl">
          Kategorija trenutno nije dostupna
        </h1>
        <p className="mt-3 text-sm text-stone-500">Pokušajte ponovo.</p>
        <Link
          href="/proizvodi"
          className="mt-6 inline-flex rounded-full bg-stone-900 px-5 py-2.5 text-sm text-white"
        >
          Proizvodi
        </Link>
      </div>
    );
  }

  const [groups, result] = await Promise.all([
    fetchPublicProductGroups({ categorySlug: slug }).catch(() => []),
    safeFetchPublicProducts({
      page,
      limit: 20,
      categorySlug: slug,
      groupSlug: group || undefined,
      search: search || undefined,
      onSale: onSale || undefined,
      inStock: inStock || undefined,
      sort,
    }),
  ]);

  const products = result?.products ?? [];
  const basePath = `/kategorije/${slug}`;

  function makeHref(nextPage: number) {
    const q = new URLSearchParams();
    if (search) q.set("search", search);
    if (group) q.set("group", group);
    if (onSale) q.set("onSale", "true");
    if (inStock) q.set("inStock", "true");
    if (sort && sort !== "recommended") q.set("sort", sort);
    if (nextPage > 1) q.set("page", String(nextPage));
    const qs = q.toString();
    return qs ? `${basePath}?${qs}` : basePath;
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <StorefrontBreadcrumb
        items={[
          { label: "Početna", href: "/" },
          { label: category.name },
        ]}
      />

      <div className="mb-8 max-w-2xl">
        <p className="text-xs uppercase tracking-[0.16em] text-stone-400">
          Kategorija
        </p>
        <h1 className="mt-2 font-[family-name:var(--font-storefront-display)] text-4xl text-stone-900">
          {category.name}
        </h1>
      </div>

      {groups.length > 0 ? (
        <div className="mb-8 flex gap-2 overflow-x-auto pb-1">
          <Link
            href={basePath}
            className={`shrink-0 rounded-full border px-4 py-2 text-sm transition ${
              !group
                ? "border-stone-900 bg-stone-900 text-white"
                : "border-stone-200 bg-white text-stone-700 hover:border-stone-300"
            }`}
          >
            Sve
          </Link>
          {groups.map((item) => (
            <Link
              key={item.id}
              href={`${basePath}?group=${encodeURIComponent(item.slug)}`}
              className={`shrink-0 rounded-full border px-4 py-2 text-sm transition ${
                group === item.slug
                  ? "border-stone-900 bg-stone-900 text-white"
                  : "border-stone-200 bg-white text-stone-700 hover:border-stone-300"
              }`}
            >
              {item.name}
            </Link>
          ))}
        </div>
      ) : null}

      <div className="grid gap-8 lg:grid-cols-[260px_minmax(0,1fr)]">
        <Suspense fallback={<div className="h-40 animate-pulse rounded-2xl bg-stone-100" />}>
          <CatalogFilters
            categories={[category]}
            groups={groups}
            basePath={basePath}
            lockCategorySlug={slug}
          />
        </Suspense>

        <div>
          {result == null ? (
            <StorefrontEmpty
              title="Kataloški servis nije dostupan"
              description="Pokušajte ponovo za nekoliko trenutaka."
              actionHref={basePath}
              actionLabel="Pokušajte ponovo"
            />
          ) : products.length === 0 ? (
            <StorefrontEmpty
              title="Nema proizvoda u ovoj kategoriji"
              description="Pokušajte drugu grupu ili pogledajte cijeli katalog."
              actionHref="/proizvodi"
              actionLabel="Svi proizvodi"
            />
          ) : (
            <>
              <PublicProductGrid products={products} />
              {result.pagination ? (
                <CatalogPagination
                  page={result.pagination.page}
                  totalPages={result.pagination.totalPages}
                  makeHref={makeHref}
                />
              ) : null}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
