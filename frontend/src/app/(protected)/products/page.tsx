"use client";

import { Suspense } from "react";

import { ProductsWorkspace } from "@/components/products/ProductsWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function ProductsPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="h-16 animate-pulse rounded-2xl bg-stone-100" />
          <div className="h-28 animate-pulse rounded-2xl bg-stone-100" />
          <ListSkeleton rows={5} />
        </div>
      }
    >
      <ProductsWorkspace />
    </Suspense>
  );
}
