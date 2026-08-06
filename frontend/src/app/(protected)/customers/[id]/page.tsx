"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { CustomerDetailsView } from "@/components/customers/CustomerDetails";

export default function CustomerDetailPage() {
  const params = useParams();
  const id = Number(params.id);

  if (!Number.isFinite(id) || id <= 0) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-red-700">Neispravan ID kupca.</p>
        <Link href="/customers" className="text-sm font-medium text-[#8a6a45]">
          Nazad na listu
        </Link>
      </div>
    );
  }

  return <CustomerDetailsView customerId={id} />;
}
