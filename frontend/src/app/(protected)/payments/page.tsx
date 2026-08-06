"use client";

import { Suspense } from "react";

import { PaymentsWorkspace } from "@/components/payments/PaymentsWorkspace";
import { ListSkeleton } from "@/components/ui/EmptyState";

export default function PaymentsPage() {
  return (
    <Suspense fallback={<ListSkeleton rows={4} />}>
      <PaymentsWorkspace />
    </Suspense>
  );
}
