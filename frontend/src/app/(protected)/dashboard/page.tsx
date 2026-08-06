"use client";

import { DashboardHeader, canViewFinance } from "@/components/dashboard/DashboardHeader";
import { FinanceSummarySection } from "@/components/dashboard/FinanceSummarySection";
import { LowStockSection } from "@/components/dashboard/LowStockSection";
import { QuickActionsSection } from "@/components/dashboard/QuickActionsSection";
import { RecentInvoicesSection } from "@/components/dashboard/RecentInvoicesSection";
import { WorkerTodayStat } from "@/components/dashboard/WorkerTodayStat";
import { useAuth } from "@/components/auth/AuthProvider";
import { todayLocalISODate } from "@/lib/dashboard";

export default function DashboardPage() {
  const { user } = useAuth();
  const today = todayLocalISODate();

  if (!user) {
    return null;
  }

  const showFinance = canViewFinance(user.role);

  return (
    <div className="space-y-5 pb-4">
      <DashboardHeader
        username={user.username}
        role={user.role}
        date={today}
      />

      <QuickActionsSection />

      {showFinance ? (
        <FinanceSummarySection date={today} />
      ) : (
        <WorkerTodayStat date={today} />
      )}

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-2">
        <LowStockSection />
        <RecentInvoicesSection />
      </div>
    </div>
  );
}
