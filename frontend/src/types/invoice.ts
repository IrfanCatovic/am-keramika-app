export type InvoiceStatus =
  | "paid"
  | "unpaid"
  | "partially_paid"
  | "cancelled";

export interface InvoiceCustomer {
  id: number;
  name: string;
  phone: string;
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

export interface PaginatedInvoiceResponse {
  data: InvoiceListItem[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}
