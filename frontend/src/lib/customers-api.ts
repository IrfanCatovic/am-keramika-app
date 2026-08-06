import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CreateCustomerPayload,
  CustomerDetails,
  CustomerFinancialSummary,
  CustomerListParams,
  CustomerMutationResponse,
  CustomerOpenInvoice,
  CustomerPayment,
  PaginatedCustomers,
  UpdateCustomerPayload,
} from "@/types/customer";

export { getApiBusinessMessage };

function buildListQuery(params: CustomerListParams): string {
  const searchParams = new URLSearchParams();
  if (params.page && params.page > 0) {
    searchParams.set("page", String(params.page));
  }
  if (params.limit && params.limit > 0) {
    searchParams.set("limit", String(params.limit));
  }
  if (params.search?.trim()) {
    searchParams.set("search", params.search.trim());
  }
  if (params.includeInactive) {
    searchParams.set("includeInactive", "true");
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export async function fetchCustomers(
  params: CustomerListParams = {},
): Promise<PaginatedCustomers> {
  return apiRequest<PaginatedCustomers>(`/customers${buildListQuery(params)}`);
}

/** Pretraga aktivnih kupaca za selector (invoice forme kasnije). */
export async function searchActiveCustomers(
  search: string,
  limit = 20,
): Promise<PaginatedCustomers> {
  return fetchCustomers({
    page: 1,
    limit,
    search,
    includeInactive: false,
  });
}

export async function fetchCustomer(id: number): Promise<CustomerDetails> {
  return apiRequest<CustomerDetails>(`/customers/${id}`);
}

export async function createCustomer(
  payload: CreateCustomerPayload,
): Promise<CustomerMutationResponse> {
  return apiRequest<CustomerMutationResponse>("/customers", {
    method: "POST",
    body: payload,
  });
}

export async function updateCustomer(
  id: number,
  payload: UpdateCustomerPayload,
): Promise<CustomerMutationResponse> {
  return apiRequest<CustomerMutationResponse>(`/customers/${id}`, {
    method: "PUT",
    body: payload,
  });
}

export async function updateCustomerStatus(
  id: number,
  isActive: boolean,
): Promise<CustomerMutationResponse> {
  return apiRequest<CustomerMutationResponse>(`/customers/${id}/status`, {
    method: "PUT",
    body: { isActive },
  });
}

export async function deleteCustomer(id: number): Promise<void> {
  await apiRequest<{ message: string }>(`/customers/${id}`, {
    method: "DELETE",
  });
}

export async function fetchCustomerOpenInvoices(
  id: number,
): Promise<CustomerOpenInvoice[]> {
  const response = await apiRequest<{
    data: CustomerOpenInvoice[];
    message?: string;
  }>(`/customers/${id}/open-invoices`);
  return response.data ?? [];
}

export async function fetchCustomerPayments(
  id: number,
): Promise<CustomerPayment[]> {
  const response = await apiRequest<{ data: CustomerPayment[] }>(
    `/customers/${id}/payments`,
  );
  return response.data ?? [];
}

export async function fetchCustomerFinancialSummary(
  id: number,
): Promise<CustomerFinancialSummary> {
  const response = await apiRequest<{
    data: CustomerFinancialSummary;
    message?: string;
  }>(`/customers/${id}/financial-summary`);
  return response.data;
}
