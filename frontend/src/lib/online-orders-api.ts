import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  ConfirmRequest,
  ConfirmResponse,
  OnlineOrderDetail,
  OnlineOrderListParams,
  OnlineOrderListResponse,
  PendingCount,
} from "@/types/online-order-staff";

export { getApiBusinessMessage };

export const ONLINE_ORDERS_CHANGED_EVENT = "online-orders-changed";

export function notifyOnlineOrdersChanged(): void {
  if (typeof window === "undefined") {
    return;
  }
  window.dispatchEvent(new Event(ONLINE_ORDERS_CHANGED_EVENT));
}

function buildListQuery(params: OnlineOrderListParams): string {
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
  if (params.search?.trim()) {
    searchParams.set("search", params.search.trim());
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

export async function fetchPendingCount(): Promise<PendingCount> {
  return apiRequest<PendingCount>("/online-orders/pending-count");
}

export async function fetchOnlineOrders(
  params: OnlineOrderListParams = {},
): Promise<OnlineOrderListResponse> {
  return apiRequest<OnlineOrderListResponse>(
    `/online-orders${buildListQuery(params)}`,
  );
}

export async function fetchOnlineOrderById(
  id: number,
): Promise<OnlineOrderDetail> {
  return apiRequest<OnlineOrderDetail>(`/online-orders/${id}`);
}

export async function confirmOnlineOrder(
  id: number,
  body: ConfirmRequest,
): Promise<ConfirmResponse> {
  return apiRequest<ConfirmResponse>(`/online-orders/${id}/confirm`, {
    method: "POST",
    body,
  });
}

export async function deleteOnlineOrder(id: number): Promise<void> {
  await apiRequest<void>(`/online-orders/${id}`, {
    method: "DELETE",
  });
}

export function onlineOrderStatusLabel(status: string): string {
  if (status === "pending") {
    return "Na čekanju";
  }
  if (status === "confirmed") {
    return "Potvrđena";
  }
  return status;
}

export function onlineOrderCustomerName(order: {
  firstName: string;
  lastName: string;
}): string {
  return `${order.firstName} ${order.lastName}`.trim();
}

/** Relative time: Prije X min / Prije X h / Prije X d */
export function formatRelativeReceived(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const diffMs = Math.max(0, Date.now() - date.getTime());
  const mins = Math.floor(diffMs / 60_000);
  if (mins < 60) {
    return `Prije ${mins} min`;
  }
  const hours = Math.floor(mins / 60);
  if (hours < 24) {
    return `Prije ${hours} h`;
  }
  const days = Math.floor(hours / 24);
  return `Prije ${days} d`;
}

export function formatOrderDateTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return iso;
  }
  return date.toLocaleString("sr-RS", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
