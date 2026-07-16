package repositories

import (
	"time"
	"am-keramika-backend/dto"
	"am-keramika-backend/database"
	"am-keramika-backend/models"
)

func GetDailyReport(startDate time.Time, endDate time.Time)(*dto.DailyReportResponse, error){

	var paymentStats struct {
		Total float64 `gorm:"column:total"`
		Count int64 `gorm:"column:count"`
	}

	err := database.DB.Model(&models.Payment{}).
	Select("COALESCE(SUM(total_amount), 0) AS total, COUNT(*) AS count",).
	Where("created_at >= ? AND created_at < ?", startDate, endDate).
	Scan(&paymentStats).Error

	if err != nil {
		return nil, err
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
		return nil, err
	}

	netCash := paymentStats.Total - refundStats.Total
	response := &dto.DailyReportResponse{
		Date: startDate.Format("2006-01-02"),
		TotalPayments: paymentStats.Total,
		TotalRefunds: refundStats.Total,
		NetCash: netCash,
		PaymentsCount: paymentStats.Count,
		RefundsCount: refundStats.Count,
	}

	return response, nil
}