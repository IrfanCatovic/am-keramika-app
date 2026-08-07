"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, useEffect, useMemo, useState, useTransition } from "react";

import type {
  PublicCategory,
  PublicProductGroup,
  PublicProductSort,
} from "@/types/public-catalog";

const SORT_OPTIONS: { value: PublicProductSort; label: string }[] = [
  { value: "recommended", label: "Preporučeno" },
  { value: "price_asc", label: "Cijena: niža prvo" },
  { value: "price_desc", label: "Cijena: viša prvo" },
  { value: "name_asc", label: "Naziv A–Z" },
];

function buildCatalogHref(params: Record<string, string | undefined>) {
  const q = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value) q.set(key, value);
  }
  const qs = q.toString();
  return qs ? `/proizvodi?${qs}` : "/proizvodi";
}

export function CatalogFilters({
  categories,
  groups,
  basePath = "/proizvodi",
  lockCategorySlug,
}: {
  categories: PublicCategory[];
  groups: PublicProductGroup[];
  basePath?: string;
  lockCategorySlug?: string;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [pending, startTransition] = useTransition();
  const [filtersOpen, setFiltersOpen] = useState(false);

  const search = searchParams.get("search") ?? "";
  const category = lockCategorySlug ?? searchParams.get("category") ?? "";
  const group = searchParams.get("group") ?? "";
  const onSale = searchParams.get("onSale") === "true";
  const inStock = searchParams.get("inStock") === "true";
  const sort = (searchParams.get("sort") as PublicProductSort) || "recommended";

  const filteredGroups = useMemo(() => {
    if (!category) return groups;
    const cat = categories.find((c) => c.slug === category);
    if (!cat) return groups;
    return groups.filter((g) => g.categoryID === cat.id);
  }, [groups, categories, category]);

  function push(patch: Record<string, string | undefined>) {
    const next = {
      search: search || undefined,
      category: lockCategorySlug ? lockCategorySlug : category || undefined,
      group: group || undefined,
      onSale: onSale ? "true" : undefined,
      inStock: inStock ? "true" : undefined,
      sort: sort !== "recommended" ? sort : undefined,
      ...patch,
    };
    if (lockCategorySlug) {
      const q = new URLSearchParams();
      if (next.search) q.set("search", next.search);
      if (next.group) q.set("group", next.group);
      if (next.onSale) q.set("onSale", next.onSale);
      if (next.inStock) q.set("inStock", next.inStock);
      if (next.sort) q.set("sort", next.sort);
      if (patch.page) q.set("page", patch.page);
      const qs = q.toString();
      startTransition(() => {
        router.push(qs ? `${basePath}?${qs}` : basePath);
        setFiltersOpen(false);
      });
      return;
    }
    startTransition(() => {
      router.push(buildCatalogHref(next));
      setFiltersOpen(false);
    });
  }

  return (
    <div className={pending ? "opacity-70 transition" : ""}>
      <CatalogSearchField
        key={search}
        initialSearch={search}
        onDebouncedSearch={(value) => {
          if (value.trim() === search.trim()) return;
          push({ search: value.trim() || undefined, page: undefined });
        }}
        onSubmitSearch={(value) => {
          push({ search: value.trim() || undefined, page: undefined });
        }}
      />

      <div className="mb-4 lg:hidden">
        <button
          type="button"
          onClick={() => setFiltersOpen(true)}
          className="inline-flex min-h-10 items-center rounded-full border border-stone-200 bg-white px-4 text-sm font-medium text-stone-800"
        >
          Filteri
        </button>
      </div>

      <aside className="hidden rounded-2xl border border-stone-200 bg-white p-4 lg:block">
        <p className="mb-4 font-[family-name:var(--font-storefront-display)] text-lg text-stone-900">
          Filteri
        </p>
        <FilterFields
          categories={categories}
          filteredGroups={filteredGroups}
          category={category}
          group={group}
          onSale={onSale}
          inStock={inStock}
          sort={sort}
          lockCategorySlug={lockCategorySlug}
          basePath={basePath}
          push={push}
        />
      </aside>

      {filtersOpen ? (
        <div className="fixed inset-0 z-50 lg:hidden" role="dialog" aria-modal>
          <button
            type="button"
            className="absolute inset-0 bg-stone-900/40"
            aria-label="Zatvori filtere"
            onClick={() => setFiltersOpen(false)}
          />
          <div className="absolute inset-x-0 bottom-0 max-h-[85vh] overflow-y-auto rounded-t-3xl bg-[#f7f5f2] p-5 shadow-xl">
            <div className="mb-4 flex items-center justify-between">
              <p className="font-[family-name:var(--font-storefront-display)] text-lg">
                Filteri
              </p>
              <button
                type="button"
                className="rounded-full border border-stone-200 bg-white px-3 py-1.5 text-sm"
                onClick={() => setFiltersOpen(false)}
              >
                Zatvori
              </button>
            </div>
            <FilterFields
              categories={categories}
              filteredGroups={filteredGroups}
              category={category}
              group={group}
              onSale={onSale}
              inStock={inStock}
              sort={sort}
              lockCategorySlug={lockCategorySlug}
              basePath={basePath}
              push={push}
            />
          </div>
        </div>
      ) : null}
    </div>
  );
}

function CatalogSearchField({
  initialSearch,
  onDebouncedSearch,
  onSubmitSearch,
}: {
  initialSearch: string;
  onDebouncedSearch: (value: string) => void;
  onSubmitSearch: (value: string) => void;
}) {
  const [searchInput, setSearchInput] = useState(initialSearch);

  useEffect(() => {
    const handle = window.setTimeout(() => {
      onDebouncedSearch(searchInput);
    }, 300);
    return () => window.clearTimeout(handle);
    // Parent callback is intentionally excluded to avoid re-debounce loops.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchInput]);

  return (
    <form
      onSubmit={(event: FormEvent) => {
        event.preventDefault();
        onSubmitSearch(searchInput);
      }}
      className="mb-4"
    >
      <label className="sr-only" htmlFor="catalog-search">
        Pretraga
      </label>
      <input
        id="catalog-search"
        value={searchInput}
        onChange={(e) => setSearchInput(e.target.value)}
        placeholder="Pretražite proizvode..."
        className="w-full rounded-full border border-stone-200 bg-white px-4 py-2.5 text-sm outline-none ring-[#c4a484]/30 focus:ring-2"
      />
    </form>
  );
}

function FilterFields({
  categories,
  filteredGroups,
  category,
  group,
  onSale,
  inStock,
  sort,
  lockCategorySlug,
  basePath,
  push,
}: {
  categories: PublicCategory[];
  filteredGroups: PublicProductGroup[];
  category: string;
  group: string;
  onSale: boolean;
  inStock: boolean;
  sort: PublicProductSort;
  lockCategorySlug?: string;
  basePath: string;
  push: (patch: Record<string, string | undefined>) => void;
}) {
  return (
    <div className="space-y-5">
      {!lockCategorySlug ? (
        <div>
          <label className="mb-1.5 block text-xs uppercase tracking-[0.14em] text-stone-400">
            Kategorija
          </label>
          <select
            value={category}
            onChange={(e) =>
              push({
                category: e.target.value || undefined,
                group: undefined,
                page: undefined,
              })
            }
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-[#c4a484]/40"
          >
            <option value="">Sve kategorije</option>
            {categories.map((item) => (
              <option key={item.id} value={item.slug}>
                {item.name}
              </option>
            ))}
          </select>
        </div>
      ) : null}

      <div>
        <label className="mb-1.5 block text-xs uppercase tracking-[0.14em] text-stone-400">
          Grupa
        </label>
        <select
          value={group}
          onChange={(e) =>
            push({ group: e.target.value || undefined, page: undefined })
          }
          className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-[#c4a484]/40"
          disabled={filteredGroups.length === 0}
        >
          <option value="">Sve grupe</option>
          {filteredGroups.map((item) => (
            <option key={item.id} value={item.slug}>
              {item.name}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label className="mb-1.5 block text-xs uppercase tracking-[0.14em] text-stone-400">
          Sortiranje
        </label>
        <select
          value={sort}
          onChange={(e) =>
            push({
              sort:
                e.target.value === "recommended"
                  ? undefined
                  : e.target.value,
              page: undefined,
            })
          }
          className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-[#c4a484]/40"
        >
          {SORT_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      <label className="flex cursor-pointer items-center gap-2 text-sm text-stone-700">
        <input
          type="checkbox"
          checked={onSale}
          onChange={(e) =>
            push({
              onSale: e.target.checked ? "true" : undefined,
              page: undefined,
            })
          }
          className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
        />
        Na akciji
      </label>

      <label className="flex cursor-pointer items-center gap-2 text-sm text-stone-700">
        <input
          type="checkbox"
          checked={inStock}
          onChange={(e) =>
            push({
              inStock: e.target.checked ? "true" : undefined,
              page: undefined,
            })
          }
          className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
        />
        Samo na stanju
      </label>

      <Link
        href={lockCategorySlug ? basePath : "/proizvodi"}
        className="inline-flex text-sm text-stone-500 underline-offset-2 hover:text-stone-800 hover:underline"
      >
        Poništi filtere
      </Link>
    </div>
  );
}
