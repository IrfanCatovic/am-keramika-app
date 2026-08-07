"use client";

import Image from "next/image";
import Link from "next/link";

import {
  EmptyState,
  InlineError,
  ListSkeleton,
  StatusBadge,
} from "@/components/ui/EmptyState";
import { ProductSalePrice } from "@/components/products/ProductSalePrice";
import { formatQuantity } from "@/lib/format";
import { Product, ProductPagination } from "@/types/product";

function PricingModeBadge({ mode }: { mode: string }) {
  if (mode === "calculated") {
    return (
      <span className="inline-flex items-center rounded-md bg-[#faf6f1] px-2 py-0.5 text-xs font-medium text-[#8a6a45] ring-1 ring-inset ring-[#c4a484]/50">
        Automatski
      </span>
    );
  }
  return (
    <span className="inline-flex items-center rounded-md bg-stone-100 px-2 py-0.5 text-xs font-medium text-stone-600 ring-1 ring-inset ring-stone-200">
      Ručno
    </span>
  );
}

function ProductThumb({ product }: { product: Product }) {
  if (product.primaryImage?.url) {
    return (
      <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded-xl bg-stone-100 ring-1 ring-stone-200">
        <Image
          src={product.primaryImage.url}
          alt={product.name}
          fill
          className="object-cover"
          sizes="48px"
          unoptimized
        />
      </div>
    );
  }
  return (
    <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-stone-100 text-[10px] font-semibold uppercase tracking-wide text-stone-500 ring-1 ring-stone-200">
      AM
    </div>
  );
}

function ProductCard({
  product,
  busy,
  onToggleActive,
}: {
  product: Product;
  busy: boolean;
  onToggleActive: (product: Product) => void;
}) {
  return (
    <article
      className={`dash-enter rounded-2xl border border-stone-200 bg-white p-4 ${
        product.isActive ? "" : "opacity-75"
      }`}
    >
      <div className="flex gap-3">
        <ProductThumb product={product} />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <p className="break-words text-sm font-semibold text-stone-900">
              {product.name}
            </p>
            <StatusBadge active={product.isActive} />
          </div>
          <p className="mt-1 break-words text-xs text-stone-500">
            {product.category?.name ?? "—"}
            {product.group?.name ? ` · ${product.group.name}` : " · Bez grupe"}
          </p>
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-sm text-stone-600">
            <ProductSalePrice product={product} />
            <span>
              {formatQuantity(product.stockQuantity)} {product.unit}
            </span>
            <PricingModeBadge mode={product.pricingMode} />
          </div>
        </div>
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        <Link
          href={`/products/${product.id}/edit`}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 transition hover:bg-stone-50"
        >
          Uredi
        </Link>
        <button
          type="button"
          disabled={busy}
          onClick={() => onToggleActive(product)}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
        >
          {product.isActive ? "Deaktiviraj" : "Aktiviraj"}
        </button>
      </div>
    </article>
  );
}

export function ProductList({
  products,
  pagination,
  loading,
  error,
  busyId,
  onRetry,
  onToggleActive,
  onPageChange,
}: {
  products: Product[];
  pagination: ProductPagination | null;
  loading: boolean;
  error: string | null;
  busyId: number | null;
  onRetry: () => void;
  onToggleActive: (product: Product) => void;
  onPageChange: (page: number) => void;
}) {
  return (
    <section className="min-w-0 rounded-2xl border border-stone-200/90 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="border-b border-stone-100 px-4 py-3.5 sm:px-5 sm:py-4">
        <h2 className="text-base font-semibold tracking-tight text-stone-900">
          Lista proizvoda
        </h2>
        {pagination ? (
          <p className="mt-0.5 text-sm text-stone-500">
            {pagination.totalItems} ukupno
            {pagination.totalPages > 0
              ? ` · strana ${pagination.page} / ${pagination.totalPages}`
              : ""}
          </p>
        ) : null}
      </div>

      <div className="px-4 py-4 sm:px-5">
        {loading ? <ListSkeleton rows={5} /> : null}

        {!loading && error ? (
          <InlineError message={error} onRetry={onRetry} />
        ) : null}

        {!loading && !error && products.length === 0 ? (
          <EmptyState
            title="Nema proizvoda"
            description="Podesite filtere ili dodajte novi proizvod."
            action={
              <Link
                href="/products/new"
                className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
              >
                Novi proizvod
              </Link>
            }
          />
        ) : null}

        {!loading && !error && products.length > 0 ? (
          <>
            <div className="space-y-3 md:hidden">
              {products.map((product) => (
                <ProductCard
                  key={product.id}
                  product={product}
                  busy={busyId === product.id}
                  onToggleActive={onToggleActive}
                />
              ))}
            </div>

            <div className="hidden overflow-x-auto md:block">
              <table className="min-w-full border-separate border-spacing-0 text-left text-sm">
                <thead>
                  <tr className="text-xs uppercase tracking-wide text-stone-500">
                    <th className="border-b border-stone-100 pb-3 pr-3 font-medium">
                      Proizvod
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Kategorija
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Grupa
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Jedinica
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Cijena
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Stanje
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Status
                    </th>
                    <th className="border-b border-stone-100 px-3 pb-3 font-medium">
                      Režim
                    </th>
                    <th className="border-b border-stone-100 pb-3 pl-3 font-medium">
                      Akcije
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {products.map((product) => (
                    <tr
                      key={product.id}
                      className={product.isActive ? undefined : "opacity-70"}
                    >
                      <td className="border-b border-stone-50 py-3 pr-3">
                        <div className="flex min-w-0 items-center gap-3">
                          <ProductThumb product={product} />
                          <span className="break-words font-medium text-stone-900">
                            {product.name}
                          </span>
                        </div>
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3 text-stone-600">
                        {product.category?.name ?? "—"}
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3 text-stone-600">
                        {product.group?.name ?? "—"}
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3 text-stone-600">
                        {product.unit}
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3">
                        <ProductSalePrice product={product} />
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3 tabular-nums text-stone-700">
                        {formatQuantity(product.stockQuantity)}
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3">
                        <StatusBadge active={product.isActive} />
                      </td>
                      <td className="border-b border-stone-50 px-3 py-3">
                        <PricingModeBadge mode={product.pricingMode} />
                      </td>
                      <td className="border-b border-stone-50 py-3 pl-3">
                        <div className="flex flex-wrap gap-2">
                          <Link
                            href={`/products/${product.id}/edit`}
                            className="inline-flex min-h-9 items-center rounded-lg border border-stone-200 px-2.5 text-xs font-medium text-stone-700 hover:bg-stone-50"
                          >
                            Uredi
                          </Link>
                          <button
                            type="button"
                            disabled={busyId === product.id}
                            onClick={() => onToggleActive(product)}
                            className="inline-flex min-h-9 items-center rounded-lg border border-stone-200 px-2.5 text-xs font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
                          >
                            {product.isActive ? "Deaktiviraj" : "Aktiviraj"}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        ) : null}

        {!loading &&
        !error &&
        pagination &&
        pagination.totalPages > 1 ? (
          <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-stone-100 pt-4">
            <button
              type="button"
              disabled={pagination.page <= 1}
              onClick={() => onPageChange(pagination.page - 1)}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-50"
            >
              Prethodna
            </button>
            <p className="text-sm text-stone-500">
              Strana {pagination.page} / {pagination.totalPages}
            </p>
            <button
              type="button"
              disabled={pagination.page >= pagination.totalPages}
              onClick={() => onPageChange(pagination.page + 1)}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-50"
            >
              Sljedeća
            </button>
          </div>
        ) : null}
      </div>
    </section>
  );
}
