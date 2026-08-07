"use client";

import { Suspense } from "react";

import { RequireRoles } from "@/components/auth/RequireRoles";
import { RefundsWorkspace } from "@/components/refunds/RefundsWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function RefundsPage() {
  return (
    <RequireRoles roles={["developer", "sef", "menadzer"]}>
      <Suspense
        fallback={
          <div className="space-y-4">
            <ListSkeleton rows={2} />
            <ListSkeleton rows={5} />
          </div>
        }
      >
        <RefundsWorkspace />
      </Suspense>
    </RequireRoles>
  );
}
