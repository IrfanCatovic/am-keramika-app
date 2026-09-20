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

export interface SalesReturnItemResponse {
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

export interface SalesReturnResponse {
  id: number;
  description: string;
  totalAmount: number;
  cashRefunded: boolean;
  refundID?: number | null;
  createdAt: string;
  createdByUser?: {
    id: number;
    username: string;
    fullName?: string;
  } | null;
  items: SalesReturnItemResponse[];
}
