import Link from "next/link";

import {
  PublicAvailability,
  PublicProductPrice,
} from "@/components/storefront/PublicPrice";
import type { PublicProduct } from "@/types/public-catalog";

function ProductImage({ product }: { product: PublicProduct }) {
  const url = product.primaryImage?.url;
  if (!url) {
    return (
      <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-stone-100 to-stone-200/80">
        <span className="font-[family-name:var(--font-storefront-display)] text-3xl tracking-wide text-stone-300">
          AM
        </span>
      </div>
    );
  }
  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      src={url}
      alt={product.name}
      className="h-full w-full object-contain p-3 transition duration-500 ease-out group-hover:scale-[1.04]"
      loading="lazy"
    />
  );
}

export function PublicProductCard({ product }: { product: PublicProduct }) {
  const meta = [product.category?.name, product.group?.name]
    .filter(Boolean)
    .join(" · ");

  return (
    <Link
      href={`/proizvodi/${product.slug}`}
      className="group flex h-full flex-col overflow-hidden rounded-2xl border border-stone-200/80 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)] transition duration-300 hover:-translate-y-0.5 hover:border-stone-300 hover:shadow-[0_12px_30px_rgba(28,25,23,0.08)]"
    >
      <div className="relative aspect-[4/3] overflow-hidden bg-stone-50">
        <ProductImage product={product} />
        {product.isOnSale && product.discountPercent > 0 ? (
          <span className="absolute left-3 top-3 rounded-md bg-stone-900/90 px-2 py-1 text-xs font-medium text-white backdrop-blur-sm">
            -{Math.round(product.discountPercent)}%
          </span>
        ) : null}
      </div>
      <div className="flex flex-1 flex-col gap-2 px-4 pb-4 pt-3">
        {meta ? (
          <p className="text-[11px] uppercase tracking-[0.14em] text-stone-400">
            {meta}
          </p>
        ) : null}
        <h3 className="line-clamp-2 text-[15px] font-medium leading-snug text-stone-900 transition group-hover:text-[#6b4f32]">
          {product.name}
        </h3>
        {product.unit ? (
          <p className="text-xs text-stone-500">Jedinica: {product.unit}</p>
        ) : null}
        <div className="mt-auto space-y-2 pt-1">
          <PublicProductPrice product={product} size="sm" />
          <PublicAvailability inStock={product.inStock} />
        </div>
      </div>
    </Link>
  );
}

export function PublicProductCardSkeleton() {
  return (
    <div className="overflow-hidden rounded-2xl border border-stone-200/80 bg-white">
      <div className="aspect-[4/3] animate-pulse bg-stone-100" />
      <div className="space-y-2 p-4">
        <div className="h-3 w-1/3 animate-pulse rounded bg-stone-100" />
        <div className="h-4 w-4/5 animate-pulse rounded bg-stone-100" />
        <div className="h-4 w-1/2 animate-pulse rounded bg-stone-100" />
      </div>
    </div>
  );
}

export function PublicProductGrid({
  products,
}: {
  products: PublicProduct[];
}) {
  return (
    <div className="grid grid-cols-1 gap-4 min-[420px]:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
      {products.map((product) => (
        <PublicProductCard key={product.id} product={product} />
      ))}
    </div>
  );
}
