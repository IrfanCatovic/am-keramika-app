export interface RefundListItem {
  id: number;
  invoiceID: number;
  amount: number;
  reason: string;
  createdAt: string;
  customerID?: number | null;
  customerName?: string | null;
  createdByUser?: {
    id: number;
    username: string;
  } | null;
}

export interface RefundPagination {
  page: number;
  limit: number;
  totalItems: number;
  totalPages: number;
}

export interface PaginatedRefunds {
  refunds: RefundListItem[];
  pagination: RefundPagination;
}

export interface RefundListParams {
  page?: number;
  limit?: number;
  customerID?: number;
  invoiceID?: number;
  fromDate?: string;
  toDate?: string;
}
