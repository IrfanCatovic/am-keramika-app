import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  PaginatedRefunds,
  RefundListParams,
} from "@/types/refund";

export { getApiBusinessMessage };

function buildQuery(params: RefundListParams): string {
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
  if (params.invoiceID && params.invoiceID > 0) {
    searchParams.set("invoiceID", String(params.invoiceID));
  }
  if (params.fromDate) {
    searchParams.set("fromDate", params.fromDate);
  }
  if (params.toDate) {
    searchParams.set("toDate", params.toDate);
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export async function fetchRefunds(
  params: RefundListParams = {},
): Promise<PaginatedRefunds> {
  return apiRequest<PaginatedRefunds>(`/refunds${buildQuery(params)}`);
}
