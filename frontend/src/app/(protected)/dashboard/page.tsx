"use client";

import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { FinanceSummarySection } from "@/components/dashboard/FinanceSummarySection";
import { LowStockSection } from "@/components/dashboard/LowStockSection";
import { QuickActionsSection } from "@/components/dashboard/QuickActionsSection";
import { RecentInvoicesSection } from "@/components/dashboard/RecentInvoicesSection";
import { useAuth } from "@/components/auth/AuthProvider";
import { todayLocalISODate } from "@/lib/dashboard";
import { userDisplayName } from "@/lib/user-display";

export default function DashboardPage() {
  const { user } = useAuth();
  const today = todayLocalISODate();

  if (!user) {
    return null;
  }

  const canOpenReports =
    user.role === "developer" ||
    user.role === "sef" ||
    user.role === "menadzer";

  return (
    <div className="min-w-0 space-y-4 pb-4 sm:space-y-5">
      <DashboardHeader
        username={userDisplayName(user)}
        role={user.role}
        date={today}
      />

      <QuickActionsSection />

      <FinanceSummarySection date={today} showReportsLink={canOpenReports} />

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-2">
        <LowStockSection />
        <RecentInvoicesSection />
      </div>
    </div>
  );
}
