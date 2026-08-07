export interface SalesSummaryReport {
  fromDate: string;
  toDate: string;
  totalSales: number;
  totalCollected: number;
  outstandingAmount: number;
  totalRefunds: number;
  netCash: number;
  invoicesCount: number;
}

export interface DailyReport {
  date: string;
  totalPayments: number;
  totalRefunds: number;
  netCash: number;
  paymentsCount: number;
  refundsCount: number;
}

export interface PeriodBreakdownItem {
  period: string;
  totalPayments: number;
  totalRefunds: number;
  netCash: number;
}

export interface PeriodReport {
  fromDate: string;
  toDate: string;
  totalPayments: number;
  totalRefunds: number;
  netCash: number;
  paymentsCount: number;
  refundsCount: number;
  groupBy: "day" | "month" | string;
  breakdown: PeriodBreakdownItem[];
}

export interface FinancialTransaction {
  id: number;
  type: "payment" | "refund" | string;
  amount: number;
  date: string;
  customerID: number | null;
  customerName: string | null;
  invoiceIDs: number[];
  description: string;
}

export interface FinancialTransactionsReport {
  fromDate: string;
  toDate: string;
  totalCount: number;
  transactions: FinancialTransaction[];
}

export type ReportRangePreset =
  | "today"
  | "yesterday"
  | "this-month"
  | "last-month"
  | "custom";
