"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";

export default function HomePage() {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isLoading) {
      return;
    }
    router.replace(user ? "/dashboard" : "/login");
  }, [isLoading, user, router]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100">
      <div className="rounded-lg border border-slate-200 bg-white px-6 py-4 text-sm text-slate-600 shadow-sm">
        Učitavanje aplikacije...
      </div>
    </div>
  );
}
