import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CreatePaymentPayload,
  CreatePaymentResponse,
  PaginatedPayments,
  Payment,
  PaymentDetailResponse,
  PaymentListParams,
} from "@/types/payment";

export { getApiBusinessMessage };

function buildListQuery(params: PaymentListParams): string {
  const searchParams = new URLSearchParams();
  if (params.page && params.page > 0) {
    searchParams.set("page", String(params.page));
  }
  if (params.limit && params.limit > 0) {
    searchParams.set("limit", String(params.limit));
  }
  if (params.customerID && params.customerID > 0) {
    searchParams.set("customerID", String(params.customerID));
  }
  if (params.fromDate?.trim()) {
    searchParams.set("fromDate", params.fromDate.trim());
  }
  if (params.toDate?.trim()) {
    searchParams.set("toDate", params.toDate.trim());
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export async function fetchPayments(
  params: PaymentListParams = {},
): Promise<PaginatedPayments> {
  return apiRequest<PaginatedPayments>(`/payments${buildListQuery(params)}`);
}

export async function fetchPayment(id: number): Promise<Payment> {
  const response = await apiRequest<PaymentDetailResponse>(`/payments/${id}`);
  return response.data;
}

export async function createPayment(
  payload: CreatePaymentPayload,
): Promise<Payment> {
  const response = await apiRequest<CreatePaymentResponse>("/payments", {
    method: "POST",
    body: payload,
  });
  return response.data;
}

export function paymentCustomerLabel(payment: Payment): string {
  if (payment.customer?.name?.trim()) {
    return payment.customer.name.trim();
  }
  if (payment.customerID) {
    return `Kupac #${payment.customerID}`;
  }
  return "Bez kupca";
}

export function roundMoney(value: number): number {
  return Math.round(value * 100) / 100;
}

/** FIFO raspodela na najstarije račune (createdAt ASC). */
export function autoAllocatePayments(
  totalAmount: number,
  invoices: { id: number; remainingAmount: number; createdAt: string }[],
): Map<number, number> {
  const result = new Map<number, number>();
  let remaining = roundMoney(totalAmount);
  if (remaining <= 0) {
    return result;
  }
  const sorted = [...invoices].sort((a, b) =>
    a.createdAt.localeCompare(b.createdAt),
  );
  for (const invoice of sorted) {
    if (remaining <= 0) {
      break;
    }
    const open = roundMoney(invoice.remainingAmount);
    if (open <= 0) {
      continue;
    }
    const amount = roundMoney(Math.min(open, remaining));
    if (amount > 0) {
      result.set(invoice.id, amount);
      remaining = roundMoney(remaining - amount);
    }
  }
  return result;
}
