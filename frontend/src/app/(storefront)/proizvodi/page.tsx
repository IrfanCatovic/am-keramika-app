import { Suspense } from "react";
import type { Metadata } from "next";

import {
  CatalogFilters,
  CatalogPagination,
} from "@/components/storefront/CatalogFilters";
import { PublicProductGrid } from "@/components/storefront/PublicProductCard";
import { StorefrontEmpty } from "@/components/storefront/StorefrontSections";
import {
  fetchPublicProductGroups,
  safeFetchPublicCategories,
  safeFetchPublicProducts,
} from "@/lib/public-catalog-api";
import type { PublicProductSort } from "@/types/public-catalog";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Proizvodi",
  description: "Pregledajte kompletan asortiman AM Keramika.",
};

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function first(value: string | string[] | undefined): string {
  if (Array.isArray(value)) return value[0] ?? "";
  return value ?? "";
}

export default async function CatalogPage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  const params = await searchParams;
  const search = first(params.search).trim();
  const category = first(params.category).trim();
  const group = first(params.group).trim();
  const onSale = first(params.onSale) === "true";
  const inStock = first(params.inStock) === "true";
  const sort = (first(params.sort) || "recommended") as PublicProductSort;
  const page = Math.max(1, Number(first(params.page)) || 1);

  const [categories, groups, result] = await Promise.all([
    safeFetchPublicCategories(),
    fetchPublicProductGroups(
      category ? { categorySlug: category } : undefined,
    ).catch(() => []),
    safeFetchPublicProducts({
      page,
      limit: 20,
      search: search || undefined,
      categorySlug: category || undefined,
      groupSlug: group || undefined,
      onSale: onSale || undefined,
      inStock: inStock || undefined,
      sort,
    }),
  ]);

  const products = result?.products ?? [];
  const pagination = result?.pagination;
  const hasFilters = Boolean(search || category || group || onSale || inStock);

  function makeHref(nextPage: number) {
    const q = new URLSearchParams();
    if (search) q.set("search", search);
    if (category) q.set("category", category);
    if (group) q.set("group", group);
    if (onSale) q.set("onSale", "true");
    if (inStock) q.set("inStock", "true");
    if (sort && sort !== "recommended") q.set("sort", sort);
    if (nextPage > 1) q.set("page", String(nextPage));
    const qs = q.toString();
    return qs ? `/proizvodi?${qs}` : "/proizvodi";
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <div className="mb-8 max-w-2xl">
        <p className="text-xs uppercase tracking-[0.16em] text-stone-400">
          Katalog
        </p>
        <h1 className="mt-2 font-[family-name:var(--font-storefront-display)] text-4xl text-stone-900">
          Proizvodi
        </h1>
        <p className="mt-3 text-sm text-stone-500">
          Pretražite asortiman keramike, sanitarija, grijanja i opreme.
        </p>
      </div>

      <div className="grid gap-8 lg:grid-cols-[260px_minmax(0,1fr)]">
        <Suspense fallback={<div className="h-40 animate-pulse rounded-2xl bg-stone-100" />}>
          <CatalogFilters categories={categories} groups={groups} />
        </Suspense>

        <div>
          {result == null ? (
            <StorefrontEmpty
              title="Kataloški servis nije dostupan"
              description="Pokušajte ponovo za nekoliko trenutaka."
              actionHref="/proizvodi"
              actionLabel="Pokušajte ponovo"
            />
          ) : products.length === 0 ? (
            <StorefrontEmpty
              title={
                hasFilters
                  ? "Nismo pronašli proizvode koji odgovaraju vašoj pretrazi."
                  : "Trenutno nema proizvoda"
              }
              description={
                hasFilters
                  ? "Pokušajte drugačije filtere ili poništite pretragu."
                  : "Asortiman će uskoro biti dostupan."
              }
              actionHref="/proizvodi"
              actionLabel={hasFilters ? "Poništi filtere" : undefined}
            />
          ) : (
            <>
              <p className="mb-4 text-sm text-stone-500">
                {pagination?.totalItems ?? products.length} proizvoda
              </p>
              <PublicProductGrid products={products} />
              {pagination ? (
                <CatalogPagination
                  page={pagination.page}
                  totalPages={pagination.totalPages}
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
