"use client";

import Image from "next/image";
import Link from "next/link";

import {
  SectionBody,
  SectionCard,
  SectionEmpty,
  SkeletonBlock,
} from "@/components/dashboard/SectionCard";
import { useAsyncSection } from "@/hooks/useAsyncSection";
import { fetchLowStockPreview } from "@/lib/dashboard";
import { formatQuantity } from "@/lib/format";
import { LowStockProduct } from "@/types/inventory";

function ProductThumb({ product }: { product: LowStockProduct }) {
  if (product.primaryImage?.url) {
    return (
      <div className="relative h-11 w-11 shrink-0 overflow-hidden rounded-xl bg-stone-100 ring-1 ring-stone-200 sm:h-12 sm:w-12">
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
    <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-stone-100 text-[10px] font-semibold uppercase tracking-wide text-stone-500 ring-1 ring-stone-200 sm:h-12 sm:w-12 sm:text-xs">
      AM
    </div>
  );
}

function LowStockRow({ product }: { product: LowStockProduct }) {
  return (
    <li className="min-w-0 rounded-2xl border border-transparent px-1 py-3 transition hover:border-stone-200 hover:bg-stone-50/80 sm:px-2">
      <div className="flex gap-3">
        <ProductThumb product={product} />
        <div className="min-w-0 flex-1">
          <p className="break-words font-medium leading-snug text-stone-900">
            {product.name}
          </p>

          <div className="mt-2 grid grid-cols-1 gap-1.5 text-sm text-stone-600 sm:grid-cols-3 sm:gap-2">
            <p>
              <span className="text-stone-400">Stanje:</span>{" "}
              <span className="font-medium text-stone-800">
                {formatQuantity(product.stockQuantity)} {product.unit}
              </span>
            </p>
            <p>
              <span className="text-stone-400">Min:</span>{" "}
              <span className="font-medium text-stone-800">
                {formatQuantity(product.minStockQuantity)} {product.unit}
              </span>
            </p>
            <p className="font-medium text-[#8a6a45]">
              Nedostaje{" "}
              {formatQuantity(
                product.missingQuantity ??
                  Math.max(0, product.minStockQuantity - product.stockQuantity),
              )}{" "}
              {product.unit}
            </p>
          </div>

          <p className="mt-2 break-words text-xs leading-relaxed text-stone-500">
            {product.category?.name ?? "Bez kategorije"}
            {product.group?.name ? ` · ${product.group.name}` : ""}
          </p>
        </div>
      </div>
    </li>
  );
}

function LowStockSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 3 }).map((_, index) => (
        <div key={index} className="flex gap-3">
          <SkeletonBlock className="h-11 w-11 shrink-0 sm:h-12 sm:w-12" />
          <div className="min-w-0 flex-1 space-y-2">
            <SkeletonBlock className="h-4 w-3/4 max-w-full" />
            <SkeletonBlock className="h-3 w-full" />
            <SkeletonBlock className="h-3 w-1/2 max-w-full" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function LowStockSection() {
  const { data, error, status, retry } = useAsyncSection(
    () => fetchLowStockPreview(),
    "Nije moguće učitati low-stock proizvode.",
  );

  const products = data?.products ?? [];

  return (
    <SectionCard
      title="Nisko stanje"
      description="Proizvodi ispod minimalnog stocka"
      action={
        <Link
          href="/inventory?status=low"
          className="shrink-0 text-sm font-medium text-[#8a6a45] transition hover:text-stone-900"
        >
          Prikaži sve
        </Link>
      }
    >
      <SectionBody
        status={status}
        error={error}
        onRetry={retry}
        loadingFallback={<LowStockSkeleton />}
      >
        {products.length === 0 ? (
          <SectionEmpty message="Trenutno nema proizvoda sa niskim stanjem." />
        ) : (
          <ul className="divide-y divide-stone-100">
            {products.map((product) => (
              <LowStockRow key={product.id} product={product} />
            ))}
          </ul>
        )}
      </SectionBody>
    </SectionCard>
  );
}
