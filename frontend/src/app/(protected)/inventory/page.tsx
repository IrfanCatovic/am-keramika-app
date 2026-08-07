import { Suspense } from "react";

import { InventoryWorkspace } from "@/components/inventory/InventoryWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function InventoryPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <ListSkeleton rows={2} />
          <ListSkeleton rows={5} />
        </div>
      }
    >
      <InventoryWorkspace />
    </Suspense>
  );
}
