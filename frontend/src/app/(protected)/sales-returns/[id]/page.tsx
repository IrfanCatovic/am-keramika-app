"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { RequireRoles } from "@/components/auth/RequireRoles";
import { SalesReturnDetailWorkspace } from "@/components/sales-returns/SalesReturnDetailWorkspace";

export default function SalesReturnDetailPage() {
  const params = useParams();
  const id = Number(params.id);

  if (!Number.isFinite(id) || id <= 0) {
    return (
      <RequireRoles roles={["developer", "sef", "menadzer", "radnik"]}>
        <div className="space-y-3">
          <p className="text-sm text-red-700">Neispravan ID povrata robe.</p>
          <Link
            href="/sales-returns"
            className="text-sm font-medium text-[#8a6a45]"
          >
            Nazad na listu
          </Link>
        </div>
      </RequireRoles>
    );
  }

  return (
    <RequireRoles roles={["developer", "sef", "menadzer", "radnik"]}>
      <SalesReturnDetailWorkspace salesReturnId={id} />
    </RequireRoles>
  );
}
