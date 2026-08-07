"use client";

import { Suspense } from "react";

import { OrdersWorkspace } from "@/components/orders/OrdersWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function OrdersPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="h-16 animate-pulse rounded-2xl bg-stone-100" />
          <ListSkeleton rows={5} />
        </div>
      }
    >
      <OrdersWorkspace />
    </Suspense>
  );
}
