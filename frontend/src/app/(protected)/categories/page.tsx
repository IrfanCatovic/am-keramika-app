"use client";

import { Suspense } from "react";

import { CategoriesWorkspace } from "@/components/categories/CategoriesWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function CategoriesPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="h-16 animate-pulse rounded-2xl bg-stone-100" />
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <ListSkeleton rows={4} />
            <ListSkeleton rows={3} />
          </div>
        </div>
      }
    >
      <CategoriesWorkspace />
    </Suspense>
  );
}
