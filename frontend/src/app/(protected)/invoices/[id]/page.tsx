"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { InvoiceDetailsView } from "@/components/invoices/InvoiceDetails";

export default function InvoiceDetailPage() {
  const params = useParams();
  const id = Number(params.id);

  if (!Number.isFinite(id) || id <= 0) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-red-700">Neispravan ID računa.</p>
        <Link href="/invoices" className="text-sm font-medium text-[#8a6a45]">
          Nazad na listu
        </Link>
      </div>
    );
  }

  return <InvoiceDetailsView invoiceId={id} />;
}
