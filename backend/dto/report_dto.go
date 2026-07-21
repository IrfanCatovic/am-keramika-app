package dto

type DailyReportResponse struct{
	Date string `json:"date"`
	TotalPayments float64 `json:"totalPayments"`
	TotalRefunds float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
	PaymentsCount int64 `json:"paymentsCount"`
	RefundsCount int64 `json:"refundsCount"`
}

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

type PeriodBreakdownResponse struct{
	Period string `json:"period"`
	TotalPayments float64 `json:"totalPayments"`
	TotalRefunds float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
}

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