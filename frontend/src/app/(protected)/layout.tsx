"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { Sidebar } from "@/components/layout/Sidebar";

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace("/login");
    }
  }, [isLoading, user, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100">
        <div className="rounded-lg border border-slate-200 bg-white px-6 py-4 text-sm text-slate-600 shadow-sm">
          Provjera prijave...
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100">
        <div className="rounded-lg border border-slate-200 bg-white px-6 py-4 text-sm text-slate-600 shadow-sm">
          Preusmjeravanje na prijavu...
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen bg-[#f4f2ef]">
      <Sidebar />
      <main className="flex min-h-screen flex-1 flex-col lg:ml-0">
        <div className="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:px-6 lg:px-8">
          {children}
        </div>
      </main>
    </div>
  );
}
