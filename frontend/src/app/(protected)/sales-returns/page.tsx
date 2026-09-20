import { Suspense } from "react";

import { RequireRoles } from "@/components/auth/RequireRoles";
import { SalesReturnsWorkspace } from "@/components/sales-returns/SalesReturnsWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function SalesReturnsPage() {
  return (
    <RequireRoles roles={["developer", "sef", "menadzer", "radnik"]}>
      <Suspense
        fallback={
          <div className="space-y-4">
            <ListSkeleton rows={2} />
            <ListSkeleton rows={5} />
          </div>
        }
      >
        <SalesReturnsWorkspace />
      </Suspense>
    </RequireRoles>
  );
}
