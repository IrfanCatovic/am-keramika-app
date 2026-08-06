"use client";

import { Suspense } from "react";

import { CustomersWorkspace } from "@/components/customers/CustomersWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function CustomersPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="h-16 animate-pulse rounded-2xl bg-stone-100" />
          <ListSkeleton rows={5} />
        </div>
      }
    >
      <CustomersWorkspace />
    </Suspense>
  );
}
