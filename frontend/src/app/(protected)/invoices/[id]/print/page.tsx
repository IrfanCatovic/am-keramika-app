"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { Suspense } from "react";

import { InvoicePrintView } from "@/components/invoices/InvoicePrintView";
import { ListSkeleton } from "@/components/ui/EmptyState";

function InvoicePrintPageInner() {
  const params = useParams();
  const searchParams = useSearchParams();
  const id = Number(params.id);
  const autoprint = searchParams.get("autoprint") === "1";

  if (!Number.isFinite(id) || id <= 0) {
    return (
      <div className="space-y-3 p-4">
        <p className="text-sm text-red-700">Neispravan ID računa.</p>
        <Link href="/invoices" className="text-sm font-medium text-[#8a6a45]">
          Nazad na listu
        </Link>
      </div>
    );
  }

  return <InvoicePrintView invoiceId={id} autoprint={autoprint} />;
}

export default function InvoicePrintPage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-[210mm] p-4">
          <ListSkeleton rows={4} />
        </div>
      }
    >
      <InvoicePrintPageInner />
    </Suspense>
  );
}
