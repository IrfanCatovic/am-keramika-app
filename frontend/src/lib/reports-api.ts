import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  DailyReport,
  FinancialTransactionsReport,
  PeriodReport,
  SalesSummaryReport,
} from "@/types/report";

export { getApiBusinessMessage };

function periodQuery(fromDate: string, toDate: string): string {
  const params = new URLSearchParams({ fromDate, toDate });
  return `?${params.toString()}`;
}

export async function fetchSalesSummaryReport(
  fromDate: string,
  toDate: string,
): Promise<SalesSummaryReport> {
  return apiRequest<SalesSummaryReport>(
    `/reports/sales-summary${periodQuery(fromDate, toDate)}`,
  );
}

export async function fetchDailyReport(date: string): Promise<DailyReport> {
  const params = new URLSearchParams({ date });
  return apiRequest<DailyReport>(`/reports/daily?${params}`);
}

export async function fetchPeriodReport(
  fromDate: string,
  toDate: string,
): Promise<PeriodReport> {
  return apiRequest<PeriodReport>(
    `/reports/period${periodQuery(fromDate, toDate)}`,
  );
}

export async function fetchTransactionsReport(
  fromDate: string,
  toDate: string,
): Promise<FinancialTransactionsReport> {
  return apiRequest<FinancialTransactionsReport>(
    `/reports/transactions${periodQuery(fromDate, toDate)}`,
  );
}

export function toISODate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

export function endOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0);
}

export function addDays(date: Date, days: number): Date {
  const next = new Date(date);
  next.setDate(next.getDate() + days);
  return next;
}
