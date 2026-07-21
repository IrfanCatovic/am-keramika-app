package repositories

import (
	"time"
	"am-keramika-backend/dto"
	"am-keramika-backend/database"
	"am-keramika-backend/models"

)

type financialStats struct {
	TotalPayments float64
	TotalRefunds  float64
	NetCash       float64
	PaymentsCount int64
	RefundsCount  int64
}

type periodAmountRow struct {
	Date string `gorm:"column:date"`
	Total float64 `gorm:"column:total"`
}


//pomocna funkcija za dobijanje finansijskih podataka za period
func getFinancialStatsByPeriod(startDate time.Time, endDate time.Time)(financialStats, error){
	var paymentStats struct {
		Total float64 `gorm:"column:total"`
		Count int64 `gorm:"column:count"`
	}

	err := database.DB.Model(&models.Payment{}).
	Select("COALESCE(SUM(total_amount), 0) AS total, COUNT(*) AS count",).
	Where("created_at >= ? AND created_at < ?", startDate, endDate).
	Scan(&paymentStats).Error

	if err != nil {
		return financialStats{}, err
	}

	
	var refundStats struct {
		Total float64 `gorm:"column:total"`
		Count int64 `gorm:"column:count"`
	}

	err = database.DB.Model(&models.Refund{}).
	Select("COALESCE(SUM(amount), 0) AS total, COUNT(*) AS count",).
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


//pomocna funkcija za dobijanje dnevnog breakdowna za period
func getDailyBreakdownByPeriod(startDate time.Time, endDate time.Time)([]dto.PeriodBreakdownResponse, error){
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
			Period: dateKey,
			TotalPayments: totalPayments,
			TotalRefunds: totalRefunds,
			NetCash: totalPayments - totalRefunds,
		}

		breakdown = append(breakdown, day)
	}
	return breakdown, nil
}


func GetDailyReport(startDate time.Time, endDate time.Time)(*dto.DailyReportResponse, error){
	stats, err := getFinancialStatsByPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	response := dto.DailyReportResponse{
		Date: startDate.Format("2006-01-02"),
		TotalPayments: stats.TotalPayments,
		TotalRefunds: stats.TotalRefunds,
		NetCash: stats.NetCash,
		PaymentsCount: stats.PaymentsCount,
		RefundsCount: stats.RefundsCount,
	}

	return &response, nil
}


func GetPeriodReport(startDate time.Time, endDate time.Time)(*dto.PeriodReportResponse, error){
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
		FromDate: startDate.Format("2006-01-02"),
		ToDate: displayToDate.Format("2006-01-02"),
		TotalPayments: stats.TotalPayments,
		TotalRefunds: stats.TotalRefunds,
		NetCash: stats.NetCash,
		PaymentsCount: stats.PaymentsCount,
		RefundsCount: stats.RefundsCount,
		GroupBy: groupBy,
		Breakdown: breakdown,
	}

	return &response, nil
}

func getMonthlyBreakdownByPeriod(startDate time.Time, endDate time.Time)([]dto.PeriodBreakdownResponse, error){
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

	currentMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location(),)
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