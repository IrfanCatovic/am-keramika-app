import { ProductPagination } from "@/types/product";

export type OnlineOrderStatus = "pending" | "confirmed";

export interface OnlineOrderListItem {
  id: number;
  status: OnlineOrderStatus | string;
  firstName: string;
  lastName: string;
  phone: string;
  city: string;
  totalAmount: number;
  itemsCount: number;
  createdAt: string;
  confirmedAt?: string | null;
  invoiceID?: number | null;
}

export interface OnlineOrderListResponse {
  orders: OnlineOrderListItem[];
  pagination: ProductPagination;
}

export interface OnlineOrderItemDetail {
  productID: number;
  productName: string;
  productSlug: string;
  unit: string;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
  currentProductActive?: boolean | null;
  currentInStockEnough?: boolean | null;
}

export interface OnlineOrderDetail {
  id: number;
  status: OnlineOrderStatus | string;
  firstName: string;
  lastName: string;
  phone: string;
  city: string;
  address: string;
  email: string;
  note: string;
  totalAmount: number;
  createdAt: string;
  confirmedAt?: string | null;
  invoiceID?: number | null;
  items: OnlineOrderItemDetail[];
}

export interface ConfirmRequest {
  customerID?: number | null;
  newCustomer?: {
    name: string;
    phone: string;
  } | null;
}

export interface ConfirmResponse {
  orderID: number;
  invoiceID: number;
  status: string;
}

export interface PendingCount {
  count: number;
}

export interface OnlineOrderListParams {
  page?: number;
  limit?: number;
  status?: OnlineOrderStatus | "";
  search?: string;
  fromDate?: string;
  toDate?: string;
}
