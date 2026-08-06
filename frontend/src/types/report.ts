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
