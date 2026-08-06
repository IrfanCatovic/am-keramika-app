export interface PaymentCustomer {
  id: number;
  name: string;
  phone: string;
  isActive: boolean;
}

export interface PaymentUser {
  id: number;
  username: string;
}

export interface PaymentAllocationInvoice {
  id: number;
  totalAmount: number;
  paidAmount: number;
  status: string;
}

export interface PaymentAllocation {
  id: number;
  invoiceID: number;
  amount: number;
  invoice: PaymentAllocationInvoice;
}

export interface Payment {
  id: number;
  customerID?: number | null;
  customer?: PaymentCustomer | null;
  createdByUserID: number;
  createdByUser?: PaymentUser | null;
  totalAmount: number;
  createdAt: string;
  allocations: PaymentAllocation[];
}

export interface CreatePaymentAllocationPayload {
  invoiceID: number;
  amount: number;
}

export interface CreatePaymentPayload {
  customerID: number;
  totalAmount: number;
  allocations: CreatePaymentAllocationPayload[];
}

export interface CreatePaymentResponse {
  message?: string;
  data: Payment;
}

export interface PaymentDetailResponse {
  message?: string;
  data: Payment;
}

export interface PaginatedPayments {
  data: Payment[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface PaymentListParams {
  page?: number;
  limit?: number;
  customerID?: number;
  fromDate?: string;
  toDate?: string;
}
