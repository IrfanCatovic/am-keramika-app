"use client";

import { RequireRoles } from "@/components/auth/RequireRoles";

export default function UsersPage() {
  return (
    <RequireRoles roles={["developer", "sef"]}>
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
          Korisnici
        </h1>
        <p className="text-sm text-slate-500">
          Upravljanje korisnicima sistema (developer i šef).
        </p>
        <div className="rounded-xl border border-dashed border-slate-300 bg-white px-5 py-8 text-sm text-slate-500">
          Modul će biti implementiran u narednoj fazi.
        </div>
      </div>
    </RequireRoles>
  );
}
