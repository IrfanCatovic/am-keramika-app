"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { Sidebar } from "@/components/layout/Sidebar";

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const isPrintRoute = pathname?.includes("/print") ?? false;

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace("/login");
    }
  }, [isLoading, user, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#f4f2ef]">
        <div className="rounded-xl border border-stone-200 bg-white px-6 py-4 text-sm text-stone-600 shadow-sm">
          Provjera prijave...
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#f4f2ef]">
        <div className="rounded-xl border border-stone-200 bg-white px-6 py-4 text-sm text-stone-600 shadow-sm">
          Preusmjeravanje na prijavu...
        </div>
      </div>
    );
  }

  if (isPrintRoute) {
    return <div className="min-h-screen bg-stone-200/80">{children}</div>;
  }

  return (
    <div className="min-h-screen bg-[#f4f2ef] lg:flex">
      <Sidebar />
      <main className="min-w-0 flex-1">
        <div className="mx-auto w-full max-w-6xl px-4 py-4 sm:px-6 sm:py-6 lg:px-8">
          {children}
        </div>
      </main>
    </div>
  );
}
