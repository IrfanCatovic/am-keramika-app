package dto

type DailyReportResponse struct{
	Date string `json:"date"`
	TotalPayments float64 `json:"totalPayments"`
	TotalRefudns float64 `json:"totalRefunds"`
	NetCash float64 `json:"netCash"`
	PaymentsCount int64 `json:"paymentsCount"`
	RefundsCount int64 `json:"refundsCount"`
}