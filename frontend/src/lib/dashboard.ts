import { ApiError, apiRequest } from "@/lib/api";
import { invoiceCustomerLabel } from "@/lib/invoices-api";
import { PaginatedLowStockResponse } from "@/types/inventory";
import { PaginatedInvoiceResponse } from "@/types/invoice";
import { SalesSummaryReport } from "@/types/report";

export { invoiceCustomerLabel };

export function todayLocalISODate(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return fallback;
}

export async function fetchSalesSummary(
  date: string,
): Promise<SalesSummaryReport> {
  const params = new URLSearchParams({
    fromDate: date,
    toDate: date,
  });
  return apiRequest<SalesSummaryReport>(`/reports/sales-summary?${params}`);
}

export async function fetchLowStockPreview(): Promise<PaginatedLowStockResponse> {
  const params = new URLSearchParams({
    page: "1",
    limit: "5",
  });
  return apiRequest<PaginatedLowStockResponse>(
    `/inventory/low-stock?${params}`,
  );
}

export async function fetchRecentInvoices(): Promise<PaginatedInvoiceResponse> {
  const params = new URLSearchParams({
    page: "1",
    limit: "5",
    sort: "createdAt",
    direction: "desc",
  });
  return apiRequest<PaginatedInvoiceResponse>(`/invoices?${params}`);
}

export async function fetchTodaysInvoices(
  date: string,
): Promise<PaginatedInvoiceResponse> {
  const params = new URLSearchParams({
    page: "1",
    limit: "1",
    fromDate: date,
    toDate: date,
  });
  return apiRequest<PaginatedInvoiceResponse>(`/invoices?${params}`);
}
