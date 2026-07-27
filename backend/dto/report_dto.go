package dto

import "time"

type DailyReportResponse struct{
	Date string `json:"date"`
	TotalPayments float64 `json:"totalPayments"`
	TotalRefunds float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
	PaymentsCount int64 `json:"paymentsCount"`
	RefundsCount int64 `json:"refundsCount"`
}

//ovo je grupni ceo period
type PeriodReportResponse struct{
	FromDate string `json:"fromDate"`
	ToDate string `json:"toDate"`
	TotalPayments float64 `json:"totalPayments"`
	TotalRefunds float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
	PaymentsCount int64 `json:"paymentsCount"`
	RefundsCount int64 `json:"refundsCount"`
	GroupBy string `json:"groupBy"`
	Breakdown []PeriodBreakdownResponse `json:"breakdown"`
}

//ovo je za svaki dan ili mesec posebno iz izabranog perioda
type PeriodBreakdownResponse struct{
	Period string `json:"period"`
	TotalPayments float64 `json:"totalPayments"`
	TotalRefunds float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
}

//ovo je za period ukupna prodaja 
type SalesSummaryReportResponse struct{
	FromDate string `json:"fromDate"`
	ToDate string `json:"toDate"`
	TotalSales float64 `json:"totalSales"`
	TotalCollected float64 `json:"totalCollected"`
	OutstandingAmount float64 `json:"outstandingAmount"`
	TotalRefunds float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
	InvoicesCount int64 `json:"invoicesCount"`
}

type FinancialTransactionsReportResponse struct{
	FromDate string `json:"fromDate"`
	ToDate string `json:"toDate"`
	TotalCount int64 `json:"totalCount"`
	Transactions []FinancialTransactionResponse `json:"transactions"`
}


type FinancialTransactionResponse struct{
	ID uint `json:"id"`
	Type string `json:"type"`
	Amount float64 `json:"amount"`
	Date time.Time `json:"date"`
	CustomerID *uint `json:"customerID"`
	CustomerName *string `json:"customerName"`
	InvoiceIDs []uint `json:"invoiceIDs"`
	Description string `json:"description"`
}

