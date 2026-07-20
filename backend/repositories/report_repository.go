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

tpye dailyAmountRow struct {
	Date string `gorm:"column:date"`
	Total float64 `gorm:"column:total"`
}



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

func getDailyBreakdownByPeriod(startDate time.Time, endDate time.Time)([]dailyAmountRow, error){
	var paymentRows []dailyAmountRow
	err := database.DB.Model(&models.Payment{}).
	Select(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM-DD') AS date, COALESCE(SUM(total_amount),0) AS total`).
	Where("created_at >= ? AND created_at < ?", startDate, endDate).
	Group(`TO_CHAR(created_at AT TIME ZONE 'Europe/Belgrade', 'YYYY-MM-DD')`).
	Order("date ASC").
	Scan(&paymentRows).Error

	if err != nil {
		return nil, err
	}

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

	displayToDate := endDate.AddDate(0, 0, -1) //ovo je samo da ne prikazujemo korisniku kako mi radimo sa vremenom jer kod nas prikazuje report < 31.Jul, a on je stavio 30. Jul 


	response := dto.PeriodReportResponse{
		FromDate: startDate.Format("2006-01-02"),
		ToDate: displayToDate.Format("2006-01-02"),
		TotalPayments: stats.TotalPayments,
		TotalRefunds: stats.TotalRefunds,
		NetCash: stats.NetCash,
		PaymentsCount: stats.PaymentsCount,
		RefundsCount: stats.RefundsCount,
	}

	return &response, nil
}