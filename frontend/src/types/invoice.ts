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
  fullName?: string;
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
  requestedQuantity: number;
  saleByPackage: boolean;
  packageQuantity: number;
  packageCount: number;
  unit?: string;
  originalUnitPrice?: number | null;
  unitPrice: number;
  priceOverridden?: boolean;
  totalPrice: number;
}

export interface InvoiceCancellationSummary {
  id: number;
  invoiceID: number;
  reason: string;
  debtReducedAmount: number;
  refundedAmount: number;
  createdAt: string;
  createdByUser?: InvoiceUserSummary | null;
}

export interface RefundResponse {
  id: number;
  invoiceID: number;
  amount: number;
  reason: string;
  createdAt?: string;
  createdByUser?: InvoiceUserSummary | null;
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
  cancellation?: InvoiceCancellationSummary | null;
  refund?: RefundResponse | null;
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
  /** Konačna cena po jedinici samo za ovu stavku; ne šalji ako nema popusta na kasi. */
  priceOverride?: number;
}

export interface CreateInvoicePayload {
  customerID?: number | null;
  items: CreateInvoiceItemPayload[];
  paymentMode?: "unpaid" | "full" | "partial";
  initialPaymentAmount?: number | null;
}

export interface CreateInvoiceResponse {
  invoice: InvoiceDetails;
}

export interface CancelInvoicePayload {
  reason: string;
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
  /** Trenutna effective katalog cena (bez ručne izmene na kasi). */
  salePrice: number;
  stockQuantity: number;
  imageUrl: string | null;
  quantity: number;
  saleByPackage?: boolean;
  packageQuantity?: number;
  /** Da li je uključen „Popust na kasi“ za ovu stavku. */
  priceOverrideEnabled?: boolean;
  /** Ručna konačna cena po jedinici; koristi se samo dok je checkbox uključen. */
  priceOverride?: number | null;
}
