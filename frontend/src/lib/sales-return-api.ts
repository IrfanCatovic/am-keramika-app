import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CreateSalesReturnPayload,
  SalesReturnResponse,
} from "@/types/sales-return";

export { getApiBusinessMessage };

export async function createSalesReturn(
  payload: CreateSalesReturnPayload,
): Promise<SalesReturnResponse> {
  return apiRequest<SalesReturnResponse>("/sales-returns", {
    method: "POST",
    body: payload,
  });
}
