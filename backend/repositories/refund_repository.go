package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	DefaultRefundListPage  = 1
	DefaultRefundListLimit = 20
	MaxRefundListLimit     = 100
)

type RefundListQuery struct {
	Page       int
	Limit      int
	CustomerID string
	InvoiceID  string
	FromDate   *time.Time
	ToDate     *time.Time // exclusive end
}

func GetRefundByInvoiceID(invoiceID uint) (*models.Refund, error) {
	var refund models.Refund
	err := database.DB.
		Preload("CreatedByUser").
		Where("invoice_id = ?", invoiceID).
		First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func GetInvoiceCancellationByInvoiceID(invoiceID uint) (*models.InvoiceCancellation, error) {
	var cancellation models.InvoiceCancellation
	err := database.DB.
		Preload("CreatedByUser").
		Where("invoice_id = ?", invoiceID).
		First(&cancellation).Error
	if err != nil {
		return nil, err
	}
	return &cancellation, nil
}

func buildRefundListQuery(q RefundListQuery) *gorm.DB {
	query := database.DB.Model(&models.Refund{}).
		Joins("LEFT JOIN invoices ON invoices.id = refunds.invoice_id AND invoices.deleted_at IS NULL")

	if strings.TrimSpace(q.InvoiceID) != "" {
		query = query.Where("refunds.invoice_id = ?", q.InvoiceID)
	}
	if strings.TrimSpace(q.CustomerID) != "" {
		query = query.Where("invoices.customer_id = ?", q.CustomerID)
	}
	if q.FromDate != nil {
		query = query.Where("refunds.created_at >= ?", *q.FromDate)
	}
	if q.ToDate != nil {
		query = query.Where("refunds.created_at < ?", *q.ToDate)
	}

	return query
}

func ListRefunds(q RefundListQuery) ([]models.Refund, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultRefundListPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultRefundListLimit
	}
	if q.Limit > MaxRefundListLimit {
		q.Limit = MaxRefundListLimit
	}

	var total int64
	if err := buildRefundListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var refunds []models.Refund
	offset := (q.Page - 1) * q.Limit
	err := buildRefundListQuery(q).
		Preload("CreatedByUser").
		Preload("Invoice").
		Preload("Invoice.Customer").
		Order("refunds.created_at DESC, refunds.id DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&refunds).Error
	return refunds, total, err
}

func RefundListTotalPages(total int64, limit int) int {
	if total <= 0 || limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
