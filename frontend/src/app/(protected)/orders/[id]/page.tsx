"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { OrderDetailWorkspace } from "@/components/orders/OrderDetailWorkspace";

export default function OrderDetailPage() {
  const params = useParams();
  const id = Number(params.id);

  if (!Number.isFinite(id) || id <= 0) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-red-700">Neispravan ID narudžbine.</p>
        <Link href="/orders" className="text-sm font-medium text-[#8a6a45]">
          Nazad na listu
        </Link>
      </div>
    );
  }

  return <OrderDetailWorkspace orderId={id} />;
}
