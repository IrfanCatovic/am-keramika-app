package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"time"
	"sort"
)

type financialStats struct {
	TotalPayments float64
	TotalRefunds  float64
	NetCash       float64
	PaymentsCount int64
	RefundsCount  int64
}

type periodAmountRow struct {
	Date  string  `gorm:"column:date"`
	Total float64 `gorm:"column:total"`
}

type salesStats struct {
	TotalSales        float64
	OutstandingAmount float64
	InvoicesCount     int64
}

// pomocna funkcija za dobijanje finansijskih podataka za period
func getFinancialStatsByPeriod(startDate time.Time, endDate time.Time) (financialStats, error) {
	var paymentStats struct {
		Total float64 `gorm:"column:total"`
		Count int64   `gorm:"column:count"`
	}

	err := database.DB.Model(&models.Payment{}).
		Select("COALESCE(SUM(total_amount), 0) AS total, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Scan(&paymentStats).Error

	if err != nil {
		return financialStats{}, err
	}

	var refundStats struct {
		Total float64 `gorm:"column:total"`
		Count int64   `gorm:"column:count"`
	}

	err = database.DB.Model(&models.Refund{}).
		Select("COALESCE(SUM(amount), 0) AS total, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Scan(&refundStats).Error

	if err != nil {
		return financialStats{}, err
	}

	stats := financialStats{
		TotalPayments: paymentStats.Total,
		TotalRefunds:  refundStats.Total,
		NetCash:       paymentStats.Total - refundStats.Total,
		PaymentsCount: paymentStats.Count,
		RefundsCount:  refundStats.Count,
	}
	return stats, nil
}

// pomocna funkcija za dobijanje mesecnog breakdowna za period - ako je duze od 31 dan grupisemo odgovore po mesecima
func getMonthlyBreakdownByPeriod(startDate time.Time, endDate time.Time) ([]dto.PeriodBreakdownResponse, error) {
	var paymentRows []periodAmountRow

	err := database.DB.Model(&models.Payment{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM')`).
		Select(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM') AS date, COALESCE(SUM(total_amount),0) AS total`).
		Order("date ASC").
		Scan(&paymentRows).Error
	if err != nil {
		return nil, err
	}

	var refundRows []periodAmountRow
	err = database.DB.Model(&models.Refund{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM')`).
		Select(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM') AS date, COALESCE(SUM(amount),0) AS total`).
		Order("date ASC").
		Scan(&refundRows).Error
	if err != nil {
		return nil, err
	}

	//ovako iz slice-a pravimo mapu da bude lakse da se pristupi vrednostima
	paymentTotalsByMonth := make(map[string]float64)
	for _, row := range paymentRows {
		paymentTotalsByMonth[row.Date] = row.Total
	}

	refundTotalsByMonth := make(map[string]float64)
	for _, row := range refundRows {
		refundTotalsByMonth[row.Date] = row.Total
	}

	currentMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	breakdown := []dto.PeriodBreakdownResponse{}
	for currentMonth.Before(endDate) {
		dateKey := currentMonth.Format("2006-01")

		totalPayments := paymentTotalsByMonth[dateKey]
		totalRefunds := refundTotalsByMonth[dateKey]

		month := dto.PeriodBreakdownResponse{
			Period:        dateKey,
			TotalPayments: totalPayments,
			TotalRefunds:  totalRefunds,
			NetCash:       totalPayments - totalRefunds,
		}

		breakdown = append(breakdown, month)

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}
	return breakdown, nil
}

// pomocna funkcija za dobijanje dnevnog breakdowna za period
func getDailyBreakdownByPeriod(startDate time.Time, endDate time.Time) ([]dto.PeriodBreakdownResponse, error) {
	var paymentRows []periodAmountRow
	err := database.DB.Model(&models.Payment{}).
		Select(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM-DD') AS date, COALESCE(SUM(total_amount),0) AS total`).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM-DD')`).
		Order("date ASC").
		Scan(&paymentRows).Error

	if err != nil {
		return nil, err
	}

	var refundRows []periodAmountRow

	err = database.DB.Model(&models.Refund{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM-DD')`).
		Select(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM-DD') AS date, COALESCE(SUM(amount),0) AS total`).
		Order("date ASC").
		Scan(&refundRows).Error

	if err != nil {
		return nil, err
	}

	paymentTotalsByDate := make(map[string]float64)

	for _, row := range paymentRows {
		paymentTotalsByDate[row.Date] = row.Total
	}

	refundTotalsByDate := make(map[string]float64)
	for _, row := range refundRows {
		refundTotalsByDate[row.Date] = row.Total
	}

	breakdown := []dto.PeriodBreakdownResponse{}
	for currentDate := startDate; currentDate.Before(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		dateKey := currentDate.Format("2006-01-02")

		totalPayments := paymentTotalsByDate[dateKey]
		totalRefunds := refundTotalsByDate[dateKey]

		day := dto.PeriodBreakdownResponse{
			Period:        dateKey,
			TotalPayments: totalPayments,
			TotalRefunds:  totalRefunds,
			NetCash:       totalPayments - totalRefunds,
		}

		breakdown = append(breakdown, day)
	}
	return breakdown, nil
}

// pomocna funkcija daje nam podatke o prodaji, neplacenoj sumi i broju racuna za period
func getSalesStatsByPeriod(startDate time.Time, endDate time.Time) (salesStats, error) {
	var invoiceStats struct {
		TotalSales        float64 `gorm:"column:total_sales"`
		OutstandingAmount float64 `gorm:"column:outstanding_amount"`
		Count             int64   `gorm:"column:count"`
	}

	err := database.DB.Model(&models.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0) AS total_sales, COALESCE(SUM(total_amount - paid_amount), 0) AS outstanding_amount, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ? AND status != ?", startDate, endDate, models.InvoiceStatusCancelled).
		Scan(&invoiceStats).Error
	if err != nil {
		return salesStats{}, err
	}
	stats := salesStats{
		TotalSales:        invoiceStats.TotalSales,
		OutstandingAmount: invoiceStats.OutstandingAmount,
		InvoicesCount:     invoiceStats.Count,
	}
	return stats, nil
}

func GetDailyReport(startDate time.Time, endDate time.Time) (*dto.DailyReportResponse, error) {
	stats, err := getFinancialStatsByPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	response := dto.DailyReportResponse{
		Date:          startDate.Format("2006-01-02"),
		TotalPayments: stats.TotalPayments,
		TotalRefunds:  stats.TotalRefunds,
		NetCash:       stats.NetCash,
		PaymentsCount: stats.PaymentsCount,
		RefundsCount:  stats.RefundsCount,
	}

	return &response, nil
}

// glavna funkcija za dobijanje podataka o periodu
func GetPeriodReport(startDate time.Time, endDate time.Time) (*dto.PeriodReportResponse, error) {
	stats, err := getFinancialStatsByPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var groupBy string
	if endDate.After(startDate.AddDate(0, 0, 31)) {
		groupBy = "month"
	} else {
		groupBy = "day"
	}

	breakdown := make([]dto.PeriodBreakdownResponse, 0)
	if groupBy == "day" {
		breakdown, err = getDailyBreakdownByPeriod(startDate, endDate)
	} else {
		breakdown, err = getMonthlyBreakdownByPeriod(startDate, endDate)
	}
	if err != nil {
		return nil, err
	}

	displayToDate := endDate.AddDate(0, 0, -1) //ovo je samo da ne prikazujemo korisniku kako mi radimo sa vremenom jer kod nas prikazuje report < 31.Jul, a on je stavio 30. Jul

	response := dto.PeriodReportResponse{
		FromDate:      startDate.Format("2006-01-02"),
		ToDate:        displayToDate.Format("2006-01-02"),
		TotalPayments: stats.TotalPayments,
		TotalRefunds:  stats.TotalRefunds,
		NetCash:       stats.NetCash,
		PaymentsCount: stats.PaymentsCount,
		RefundsCount:  stats.RefundsCount,
		GroupBy:       groupBy,
		Breakdown:     breakdown,
	}

	return &response, nil
}

func GetSalesSummaryReport(startDate time.Time, endDate time.Time) (*dto.SalesSummaryReportResponse, error) {

	salesStats, err := getSalesStatsByPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	financialStats, err := getFinancialStatsByPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	displayToDate := endDate.AddDate(0, 0, -1) //ovo je samo da ne prikazujemo korisniku kako mi radimo sa vremenom jer kod nas prikazuje report < 31.Jul,
	// a on je stavio 30. Jul
	response := dto.SalesSummaryReportResponse{
		FromDate:          startDate.Format("2006-01-02"),
		ToDate:            displayToDate.Format("2006-01-02"),
		TotalSales:        salesStats.TotalSales,
		TotalCollected:    financialStats.TotalPayments,
		OutstandingAmount: salesStats.OutstandingAmount,
		TotalRefunds:      financialStats.TotalRefunds,
		NetCash:           financialStats.NetCash,
		InvoicesCount:     salesStats.InvoicesCount,
	}
	return &response, nil

}

func GetFinancialTransactionsReport(startDate time.Time, endDate time.Time) (*dto.FinancialTransactionsReportResponse, error) {
	var payments []models.Payment

	err := database.DB.Model(&models.Payment{}).
	Preload("Customer").
	Preload("Allocations").
	Where("created_at >= ? AND created_at < ?", startDate, endDate).
	Find(&payments).Error

	if err != nil {
		return nil, err
	}

	var refunds []models.Refund
	err = database.DB.Model(&models.Refund{}).
	Preload("Invoice.Customer").
	Where("created_at >= ? AND created_at < ?", startDate, endDate).
	Find(&refunds).Error

	if err != nil {
		return nil, err
	}

 	transactions := make([]dto.FinancialTransactionResponse, 0)
	

	for _, payment := range payments {
		invoiceIDs := make([]uint, 0)
		for _, allocation := range payment.Allocations {
			invoiceIDs = append(invoiceIDs, allocation.InvoiceID)
		}

		var customerName *string
		if payment.Customer != nil {
			customerName = &payment.Customer.Name
		}
		transaction := dto.FinancialTransactionResponse{
			ID: payment.ID,
			Type: "payment",
			Amount: payment.Amount,
			Date: payment.CreatedAt,
			CustomerID: payment.CustomerID,
			CustomerName: customerName,
			InvoiceIDs: invoiceIDs,
			Description: "Uplata kupca",
		}
		transactions = append(transactions, transaction)
	}
	
	for _, refund := range refunds {
		invoiceIDs := []uint{refund.InvoiceID}
	
	customerID := refund.Invoice.CustomerID
	
	var customerName *string //var koja cuva ime kupca 
	if refund.Invoice.Customer != nil {
		customerName = &refund.Invoice.Customer.Name
	}
	transaction := dto.FinancialTransactionResponse{
		ID: refund.ID,
		Type: "refund",
		Amount: refund.Amount,
		Date: refund.CreatedAt,
		CustomerID: refund.Invoice.CustomerID,
		CustomerName: customerName,
		InvoiceIDs: invoiceIDs,
		Description: refund.Description,
	}

	transactions = append(transactions, transaction)
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	response := dto.FinancialTransactionsReportResponse{
		FromDate: startDate.Format("2006-01-02"),
		ToDate: endDate.AddDate(0, 0, -1).Format("2006-01-02"), //ovo radimo jer korisnik salje 31.07 a mi moramo da mu dodamo 1 dan da bi i 31. racuna, a kad vracamo odg oduzmemo 1 dan
		TotalCount: int64(len(transactions)),
		Transactions: transactions,
	}
	return &response, nil
}