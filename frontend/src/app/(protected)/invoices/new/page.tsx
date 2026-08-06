"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";

import { InvoiceForm } from "@/components/invoices/InvoiceForm";
import { ListSkeleton } from "@/components/ui/EmptyState";

function NewInvoiceContent() {
  const searchParams = useSearchParams();
  const raw = searchParams.get("customerID");
  const parsed = raw ? Number(raw) : null;
  const initialCustomerID =
    parsed && Number.isFinite(parsed) && parsed > 0 ? parsed : null;

  return <InvoiceForm initialCustomerID={initialCustomerID} />;
}

export default function NewInvoicePage() {
  return (
    <Suspense fallback={<ListSkeleton rows={4} />}>
      <NewInvoiceContent />
    </Suspense>
  );
}
