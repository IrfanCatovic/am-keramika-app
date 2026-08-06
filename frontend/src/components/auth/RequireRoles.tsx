"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { UserRole } from "@/types/auth";

export function RequireRoles({
  roles,
  children,
}: {
  roles: UserRole[];
  children: React.ReactNode;
}) {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && user && !roles.includes(user.role)) {
      router.replace("/dashboard");
    }
  }, [isLoading, user, roles, router]);

  if (isLoading || !user) {
    return null;
  }

  if (!roles.includes(user.role)) {
    return (
      <div className="rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm text-slate-600">
        Nemate dozvolu za ovu stranicu.
      </div>
    );
  }

  return children;
}
