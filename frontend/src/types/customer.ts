export interface CustomerListItem {
  id: number;
  name: string;
  phone: string;
}

export interface Customer {
  id: number;
  name: string;
  phone: string;
}

export interface CustomerInvoiceSummary {
  id: number;
  totalAmount: number;
  status: string;
  createdAt: string;
}

export interface CustomerDetails {
  id: number;
  name: string;
  phone: string;
  debt: number;
  invoices: CustomerInvoiceSummary[];
}

export interface PaginatedCustomers {
  data: CustomerListItem[];
  page: number;
  limit: number;
  total: number;
  /** Backend koristi snake_case total_pages */
  total_pages: number;
}

export interface CustomerListParams {
  page?: number;
  limit?: number;
  search?: string;
  includeInactive?: boolean;
}

export interface CreateCustomerPayload {
  name: string;
  phone?: string;
}

export interface UpdateCustomerPayload {
  name: string;
  phone?: string;
}

export interface CustomerMutationResponse {
  message?: string;
  customer: Customer;
}

export interface CustomerOpenInvoice {
  id: number;
  totalAmount: number;
  paidAmount: number;
  remainingAmount: number;
  status: string;
  createdAt: string;
}

export interface CustomerPaymentAllocationInvoice {
  id: number;
  totalAmount: number;
  paidAmount: number;
  status: string;
}

export interface CustomerPaymentAllocation {
  id: number;
  invoiceID: number;
  amount: number;
  invoice: CustomerPaymentAllocationInvoice;
}

export interface CustomerPayment {
  id: number;
  customerID?: number | null;
  createdByUserID: number;
  totalAmount: number;
  createdAt: string;
  allocations: CustomerPaymentAllocation[];
}

export interface CustomerFinancialSummary {
  id: number;
  name: string;
  phone: string;
  totalDebt: number;
  openInvoicesCount: number;
  paymentsCount: number;
}
