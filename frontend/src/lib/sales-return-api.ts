import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CreateSalesReturnPayload,
  PaginatedSalesReturns,
  SalesReturnDetail,
  SalesReturnListParams,
} from "@/types/sales-return";

export { getApiBusinessMessage };

function buildListQuery(params: SalesReturnListParams): string {
  const searchParams = new URLSearchParams();
  if (params.page && params.page > 0) {
    searchParams.set("page", String(params.page));
  }
  if (params.pageSize && params.pageSize > 0) {
    searchParams.set("pageSize", String(params.pageSize));
  }
  if (params.search?.trim()) {
    searchParams.set("search", params.search.trim());
  }
  if (params.cashRefunded !== undefined) {
    searchParams.set("cashRefunded", String(params.cashRefunded));
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export async function createSalesReturn(
  payload: CreateSalesReturnPayload,
): Promise<SalesReturnDetail> {
  return apiRequest<SalesReturnDetail>("/sales-returns", {
    method: "POST",
    body: payload,
  });
}

export async function fetchSalesReturns(
  params: SalesReturnListParams = {},
): Promise<PaginatedSalesReturns> {
  return apiRequest<PaginatedSalesReturns>(
    `/sales-returns${buildListQuery(params)}`,
  );
}

export async function fetchSalesReturn(
  id: number,
): Promise<SalesReturnDetail> {
  return apiRequest<SalesReturnDetail>(`/sales-returns/${id}`);
}
