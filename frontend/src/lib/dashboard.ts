import { ApiError, apiRequest } from "@/lib/api";
import { PaginatedLowStockResponse } from "@/types/inventory";
import { PaginatedInvoiceResponse } from "@/types/invoice";
import { SalesSummaryReport } from "@/types/report";

export function todayLocalISODate(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatMoney(amount: number): string {
  return new Intl.NumberFormat("bs-BA", {
    style: "currency",
    currency: "BAM",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatQuantity(value: number): string {
  return new Intl.NumberFormat("bs-BA", {
    maximumFractionDigits: 3,
  }).format(value);
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

export function invoiceCustomerLabel(invoice: {
  customerName?: string;
  customer: { name: string } | null;
}): string {
  if (invoice.customerName?.trim()) {
    return invoice.customerName.trim();
  }
  if (invoice.customer?.name?.trim()) {
    return invoice.customer.name.trim();
  }
  return "Gotovinska prodaja";
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
    sort: "created_at",
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
