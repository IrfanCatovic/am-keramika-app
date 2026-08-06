export type InvoiceStatus =
  | "paid"
  | "unpaid"
  | "partially_paid"
  | "cancelled";

export type InvoiceSort = "createdAt" | "totalAmount";
export type InvoiceSortDirection = "asc" | "desc";

export interface InvoiceCustomer {
  id: number;
  name: string;
  phone: string;
  isActive: boolean;
}

export interface InvoiceUserSummary {
  id: number;
  username: string;
}

export interface InvoiceListItem {
  id: number;
  customerID: number | null;
  customer: InvoiceCustomer | null;
  customerName?: string;
  totalAmount: number;
  paidAmount: number;
  remainingAmount: number;
  status: InvoiceStatus | string;
  createdAt: string;
  createdByUser?: InvoiceUserSummary | null;
}

export interface InvoiceItem {
  productID: number;
  productName: string;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
}

export interface InvoiceDetails {
  id: number;
  customerID: number | null;
  customer: InvoiceCustomer | null;
  totalAmount: number;
  paidAmount: number;
  remainingAmount: number;
  status: InvoiceStatus | string;
  createdAt: string;
  createdByUser?: InvoiceUserSummary | null;
  items: InvoiceItem[];
}

export interface PaginatedInvoiceResponse {
  data: InvoiceListItem[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface InvoiceListParams {
  page?: number;
  limit?: number;
  status?: InvoiceStatus | "";
  customerID?: number;
  fromDate?: string;
  toDate?: string;
  search?: string;
  sort?: InvoiceSort;
  direction?: InvoiceSortDirection;
}

export interface CreateInvoiceItemPayload {
  productID: number;
  quantity: number;
}

export interface CreateInvoicePayload {
  customerID?: number | null;
  items: CreateInvoiceItemPayload[];
}

export interface CreateInvoiceResponse {
  invoice: InvoiceDetails;
}

export interface CancelInvoicePayload {
  reason: string;
}

export interface RefundResponse {
  id: number;
  invoiceID: number;
  amount: number;
  reason: string;
  createdByUser?: InvoiceUserSummary | null;
}

export interface CancelInvoiceResponse {
  id: number;
  invoiceID: number;
  reason: string;
  debtReducedAmount: number;
  refundedAmount: number;
  createdByUser?: InvoiceUserSummary | null;
  refund?: RefundResponse | null;
}

export interface CancelInvoiceApiResponse {
  data: CancelInvoiceResponse;
  message?: string;
}

/** Lokalna stavka forme (preview) — nije create DTO. */
export interface InvoiceFormLine {
  productID: number;
  name: string;
  unit: string;
  salePrice: number;
  stockQuantity: number;
  imageUrl: string | null;
  quantity: number;
}
