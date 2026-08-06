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
import { fetchLowStockPreview, formatQuantity } from "@/lib/dashboard";
import { LowStockProduct } from "@/types/inventory";

function ProductThumb({ product }: { product: LowStockProduct }) {
  if (product.primaryImage?.url) {
    return (
      <div className="relative h-14 w-14 shrink-0 overflow-hidden rounded-xl bg-stone-100 ring-1 ring-stone-200">
        <Image
          src={product.primaryImage.url}
          alt={product.name}
          fill
          className="object-cover"
          sizes="56px"
          unoptimized
        />
      </div>
    );
  }

  return (
    <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-stone-100 to-stone-200 text-xs font-semibold uppercase tracking-wide text-stone-500 ring-1 ring-stone-200">
      AM
    </div>
  );
}

function LowStockRow({ product }: { product: LowStockProduct }) {
  return (
    <li className="flex gap-3 rounded-xl border border-transparent px-2 py-3 transition hover:border-stone-200 hover:bg-stone-50/80">
      <ProductThumb product={product} />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <p className="truncate font-medium text-stone-900">{product.name}</p>
          <p className="shrink-0 text-xs font-medium text-[#8a6a45]">
            Nedostaje {formatQuantity(product.missingQuantity)} {product.unit}
          </p>
        </div>
        <p className="mt-1 text-sm text-stone-600">
          Stanje: {formatQuantity(product.stockQuantity)} {product.unit}
          <span className="mx-1.5 text-stone-300">·</span>
          Min: {formatQuantity(product.minStockQuantity)} {product.unit}
        </p>
        <p className="mt-1 truncate text-xs text-stone-500">
          {product.category?.name ?? "Bez kategorije"}
          {product.group?.name ? ` · ${product.group.name}` : ""}
        </p>
      </div>
    </li>
  );
}

function LowStockSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 3 }).map((_, index) => (
        <div key={index} className="flex gap-3">
          <SkeletonBlock className="h-14 w-14 shrink-0" />
          <div className="flex-1 space-y-2">
            <SkeletonBlock className="h-4 w-2/3" />
            <SkeletonBlock className="h-3 w-1/2" />
            <SkeletonBlock className="h-3 w-1/3" />
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
          href="/inventory"
          className="text-sm font-medium text-[#8a6a45] transition hover:text-stone-900"
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
