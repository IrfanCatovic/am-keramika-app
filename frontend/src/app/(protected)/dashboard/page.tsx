"use client";

import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { roleLabel } from "@/types/auth";

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const router = useRouter();

  if (!user) {
    return null;
  }

  function handleLogout() {
    logout();
    router.replace("/login");
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
          Dashboard
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          Početni pregled interne aplikacije AM Keramika.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <p className="text-lg font-medium text-slate-900">
          Dobrodošli, {user.username}
        </p>
        <p className="mt-2 text-sm text-slate-600">
          Uloga: <span className="font-medium">{roleLabel(user.role)}</span>
        </p>
        <p className="mt-1 text-sm text-slate-600">
          Status: {user.isActive === false ? "Neaktivan" : "Prijavljeni ste"}
        </p>
        <button
          type="button"
          onClick={handleLogout}
          className="mt-5 rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
        >
          Odjavi se
        </button>
      </div>
    </div>
  );
}
