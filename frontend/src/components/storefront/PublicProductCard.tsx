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
      <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-stone-100 via-[#f3efe9] to-stone-200/70">
        <span className="font-[family-name:var(--font-storefront-display)] text-3xl tracking-[0.12em] text-stone-300">
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
      className="h-full w-full object-contain p-4 transition duration-500 ease-out group-hover:scale-[1.03]"
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
      className="group flex h-full flex-col overflow-hidden rounded-xl border border-stone-200/90 bg-white transition duration-300 hover:-translate-y-0.5 hover:border-stone-300 hover:shadow-[0_14px_34px_rgba(28,25,23,0.07)]"
    >
      <div className="relative aspect-[4/3] overflow-hidden bg-[#f7f5f2]">
        <ProductImage product={product} />
        {product.isOnSale && product.discountPercent > 0 ? (
          <span className="absolute left-3 top-3 rounded-md bg-[#2a2420]/92 px-2 py-1 text-[11px] font-medium tracking-wide text-[#e8d5bc]">
            -{Math.round(product.discountPercent)}%
          </span>
        ) : null}
        <PublicAvailability
          inStock={product.inStock}
          className="absolute right-3 top-3 rounded-md bg-white/90 px-2 py-1 shadow-sm backdrop-blur-sm"
        />
      </div>
      <div className="flex flex-1 flex-col gap-2 px-4 pb-4 pt-3">
        {meta ? (
          <p className="text-[10px] uppercase tracking-[0.16em] text-stone-400">
            {meta}
          </p>
        ) : null}
        <h3 className="line-clamp-2 text-[15px] font-medium leading-snug text-stone-900 transition group-hover:text-[#5c4630]">
          {product.name}
        </h3>
        <div className="mt-auto pt-2">
          <PublicProductPrice product={product} size="sm" />
        </div>
      </div>
    </Link>
  );
}

export function PublicProductCardSkeleton() {
  return (
    <div className="overflow-hidden rounded-xl border border-stone-200/90 bg-white">
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
