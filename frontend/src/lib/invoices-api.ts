import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CancelInvoiceApiResponse,
  CancelInvoicePayload,
  CancelInvoiceResponse,
  CreateInvoicePayload,
  CreateInvoiceResponse,
  InvoiceDetails,
  InvoiceListParams,
  PaginatedInvoiceResponse,
} from "@/types/invoice";

export { getApiBusinessMessage };

function buildListQuery(params: InvoiceListParams): string {
  const searchParams = new URLSearchParams();
  if (params.page && params.page > 0) {
    searchParams.set("page", String(params.page));
  }
  if (params.limit && params.limit > 0) {
    searchParams.set("limit", String(params.limit));
  }
  if (params.status) {
    searchParams.set("status", params.status);
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
  if (params.search?.trim()) {
    searchParams.set("search", params.search.trim());
  }
  if (params.sort) {
    searchParams.set("sort", params.sort);
  }
  if (params.direction) {
    searchParams.set("direction", params.direction);
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export async function fetchInvoices(
  params: InvoiceListParams = {},
): Promise<PaginatedInvoiceResponse> {
  return apiRequest<PaginatedInvoiceResponse>(
    `/invoices${buildListQuery(params)}`,
  );
}

export async function fetchInvoice(id: number): Promise<InvoiceDetails> {
  return apiRequest<InvoiceDetails>(`/invoices/${id}`);
}

export async function createInvoice(
  payload: CreateInvoicePayload,
): Promise<InvoiceDetails> {
  const body: {
    customerID?: number;
    items: CreateInvoicePayload["items"];
  } = {
    items: payload.items,
  };
  if (payload.customerID != null && payload.customerID > 0) {
    body.customerID = payload.customerID;
  }

  const response = await apiRequest<CreateInvoiceResponse>("/invoices", {
    method: "POST",
    body,
  });
  return response.invoice;
}

export async function cancelInvoice(
  id: number,
  payload: CancelInvoicePayload,
): Promise<CancelInvoiceResponse> {
  const response = await apiRequest<CancelInvoiceApiResponse>(
    `/invoices/${id}/cancel`,
    {
      method: "PUT",
      body: payload,
    },
  );
  return response.data;
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
