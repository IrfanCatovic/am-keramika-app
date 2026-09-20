export interface SalesReturnCreatedByUser {
  id: number;
  username: string;
  fullName?: string;
}

export interface CreateSalesReturnItemPayload {
  productID: number;
  quantity: number;
  unitPrice: number;
}

export interface CreateSalesReturnPayload {
  description?: string;
  cashRefunded: boolean;
  items: CreateSalesReturnItemPayload[];
}

export interface SalesReturnItem {
  id: number;
  productID: number;
  productName: string;
  unit: string;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
  saleByPackage: boolean;
  packageQuantity: number;
}

/** @deprecated Prefer SalesReturnItem */
export type SalesReturnItemResponse = SalesReturnItem;

export interface SalesReturnListItem {
  id: number;
  description: string;
  totalAmount: number;
  cashRefunded: boolean;
  itemsCount: number;
  createdAt: string;
  createdByUser?: SalesReturnCreatedByUser | null;
}

export interface PaginatedSalesReturns {
  items: SalesReturnListItem[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

export interface SalesReturnListParams {
  page?: number;
  pageSize?: number;
  search?: string;
  cashRefunded?: boolean;
}

export interface SalesReturnDetail {
  id: number;
  description: string;
  totalAmount: number;
  cashRefunded: boolean;
  refundID?: number | null;
  createdAt: string;
  createdByUser?: SalesReturnCreatedByUser | null;
  items: SalesReturnItem[];
}

/** @deprecated Prefer SalesReturnDetail */
export type SalesReturnResponse = SalesReturnDetail;
