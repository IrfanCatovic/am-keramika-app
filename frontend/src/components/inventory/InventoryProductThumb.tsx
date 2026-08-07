"use client";

import Image from "next/image";

import { InventoryProductRow } from "@/types/inventory";

export function InventoryProductThumb({
  product,
}: {
  product: Pick<InventoryProductRow, "name" | "primaryImage">;
}) {
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
