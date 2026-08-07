package repositories

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

func CreateInvoice(req dto.CreateInvoiceRequest, createdByUserID uint) (*models.Invoice, error) {
	tx := database.DB.Begin()

	if req.CustomerID != nil {
		if err := validateCustomerForInvoiceTx(tx, *req.CustomerID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	invoiceStatus := models.InvoiceStatusPaid
	if req.CustomerID != nil {
		invoiceStatus = models.InvoiceStatusUnpaid
	}

	invoice := models.Invoice{
		CreatedByUserID: createdByUserID,
		CustomerID:      req.CustomerID,
		Status:          invoiceStatus,
		TotalAmount:     0,
		PaidAmount:      0,
	}

	err := tx.Create(&invoice).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var totalAmount float64 = 0

	for _, item := range req.Items {
		var product models.Product

		err := tx.First(&product, item.ProductID).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.New("proizvod nije pronađen")
		}

		if product.StockQuantity < item.Quantity {
			tx.Rollback()
			return nil, errors.New("nema dovoljno stoka na skladištu")
		}

		unitPrice := product.SalePrice
		totalPrice := unitPrice * item.Quantity

		invoiceItem := models.InvoiceItem{
			InvoiceID:  invoice.ID,
			ProductID:  product.ID,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: totalPrice,
		}

		err = tx.Create(&invoiceItem).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		product.StockQuantity -= item.Quantity

		err = tx.Save(&product).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// Ukupan iznos se sabira za svaki račun (sa kupcem i bez).
		totalAmount += totalPrice

		// Gotovinski račun: inventory movement po stavci (payment se kreira jednom nakon petlje).
		if req.CustomerID == nil {
			movement := models.InventoryMovement{
				ProductID:       product.ID,
				CreatedByUserID: createdByUserID,
				MovementType:    "sale",
				Quantity:        item.Quantity,
				Note:            "Prodaja kroz racun",
			}

			err = tx.Create(&movement).Error
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	invoice.TotalAmount = totalAmount

	if req.CustomerID == nil {
		payment := models.Payment{
			CustomerID:      nil,
			CreatedByUserID: createdByUserID,
			TotalAmount:     totalAmount,
		}

		err = tx.Create(&payment).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		allocation := models.PaymentAllocation{
			PaymentID: payment.ID,
			InvoiceID: invoice.ID,
			Amount:    totalAmount,
		}
		err = tx.Create(&allocation).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		invoice.PaidAmount = totalAmount
		invoice.Status = models.InvoiceStatusPaid
	} else {
		invoice.PaidAmount = 0
		invoice.Status = models.InvoiceStatusUnpaid

		err = tx.Model(&models.Customer{}).Where("id = ?", *req.CustomerID).
			Update("total_debt", gorm.Expr("total_debt + ?", totalAmount)).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	err = tx.Save(&invoice).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	err = database.DB.Preload("Items").Preload("Items.Product").First(&invoice, invoice.ID).Error
	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func GetInvoiceByID(id uint) (*models.Invoice, error) {

	var invoice models.Invoice

	err := database.DB.
		Preload("Customer").
		Preload("CreatedByUser").
		Preload("Items").
		Preload("Items.Product").
		First(&invoice, id).Error

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

type InvoiceListQuery struct {
	Page       int
	Limit      int
	Search     string
	Status     string
	CustomerID string
	FromDate   *time.Time
	ToDate     *time.Time // exclusive end of selected day
	Sort       string
	Direction  string
}

func buildInvoiceListQuery(q InvoiceListQuery) *gorm.DB {
	query := database.DB.Model(&models.Invoice{})

	if q.Search != "" {
		search := strings.TrimSpace(q.Search)
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Joins("LEFT JOIN customers ON customers.id = invoices.customer_id AND customers.deleted_at IS NULL")
		if invoiceID, err := strconv.ParseUint(search, 10, 64); err == nil {
			query = query.Where("invoices.id = ? OR LOWER(customers.name) LIKE ?", invoiceID, pattern)
		} else {
			query = query.Where("LOWER(customers.name) LIKE ?", pattern)
		}
	}

	if q.Status != "" {
		query = query.Where("invoices.status = ?", q.Status)
	}
	if q.CustomerID != "" {
		query = query.Where("invoices.customer_id = ?", q.CustomerID)
	}
	if q.FromDate != nil {
		query = query.Where("invoices.created_at >= ?", *q.FromDate)
	}
	if q.ToDate != nil {
		query = query.Where("invoices.created_at < ?", *q.ToDate)
	}

	return query
}

func GetAllInvoices(q InvoiceListQuery) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}

	countQuery := buildInvoiceListQuery(q)
	if err := countQuery.Distinct("invoices.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortColumn := "invoices.created_at"
	sortDirection := "DESC"
	if q.Sort == "totalAmount" {
		sortColumn = "invoices.total_amount"
	}
	if q.Direction == "asc" {
		sortDirection = "ASC"
	}

	offset := (q.Page - 1) * q.Limit
	err := buildInvoiceListQuery(q).
		Preload("Customer").
		Preload("CreatedByUser").
		Order(sortColumn + " " + sortDirection).
		Limit(q.Limit).
		Offset(offset).
		Find(&invoices).Error
	if err != nil {
		return nil, 0, err
	}
	return invoices, total, nil
}

func GetOpenInvoicesByCustomerID(customerID uint) ([]models.Invoice, error) {
	var customer models.Customer

	err := database.DB.First(&customer, customerID).Error
	if err != nil {
		return nil, errors.New("kupac nije pronađen")
	}

	var invoices []models.Invoice

	err = database.DB.Where("customer_id = ? AND status IN ?", customerID, []models.InvoiceStatus{
		models.InvoiceStatusUnpaid,
		models.InvoiceStatusPartiallyPaid,
	}).Order("created_at DESC").Find(&invoices).Error

	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func CancelInvoice(id uint, req dto.CancelInvoiceRequest, createdByUserID uint) (*dto.CancelInvoiceResponse, error) {
	var invoice models.Invoice

	tx := database.DB.Begin()

	err := tx.Model(&models.Invoice{}).Preload("Items").First(&invoice, id).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if invoice.Status == models.InvoiceStatusCancelled {
		tx.Rollback()
		return nil, errors.New("racun je vec otkazan")
	}

	remainingAmount := invoice.TotalAmount - invoice.PaidAmount
	refundedAmount := invoice.PaidAmount

	for _, item := range invoice.Items {
		var product models.Product
		err = tx.First(&product, item.ProductID).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		product.StockQuantity += item.Quantity
		err = tx.Save(&product).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		inventoryMovement := models.InventoryMovement{
			ProductID:       item.ProductID,
			CreatedByUserID: createdByUserID,
			MovementType:    "return",
			Quantity:        item.Quantity,
			Note:            "Otkazan racun",
		}
		err = tx.Create(&inventoryMovement).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if invoice.CustomerID != nil {
		var customer models.Customer
		err = tx.First(&customer, *invoice.CustomerID).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		customer.TotalDebt -= remainingAmount
		err = tx.Save(&customer).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	var refundID uint
	if refundedAmount > 0 {
		refund := models.Refund{
			InvoiceID:       invoice.ID,
			CreatedByUserID: createdByUserID,
			Amount:          refundedAmount,
			Reason:          req.Reason,
		}
		err = tx.Create(&refund).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		refundID = refund.ID
	}

	invoiceCancellation := models.InvoiceCancellation{
		InvoiceID:         invoice.ID,
		CreatedByUserID:   createdByUserID,
		Reason:            req.Reason,
		DebtReducedAmount: remainingAmount,
		RefundedAmount:    refundedAmount,
	}
	err = tx.Create(&invoiceCancellation).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	invoice.Status = models.InvoiceStatusCancelled
	err = tx.Save(&invoice).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	err = database.DB.Preload("CreatedByUser").First(&invoiceCancellation, invoiceCancellation.ID).Error
	if err != nil {
		return nil, err
	}

	response := &dto.CancelInvoiceResponse{
		ID:                invoiceCancellation.ID,
		InvoiceID:         invoiceCancellation.InvoiceID,
		Reason:            invoiceCancellation.Reason,
		DebtReducedAmount: invoiceCancellation.DebtReducedAmount,
		RefundedAmount:    invoiceCancellation.RefundedAmount,
		CreatedByUser: &dto.UserSummaryResponse{
			ID:       invoiceCancellation.CreatedByUser.ID,
			Username: invoiceCancellation.CreatedByUser.Username,
		},
	}

	if refundID > 0 {
		var refund models.Refund
		err = database.DB.Preload("CreatedByUser").First(&refund, refundID).Error
		if err != nil {
			return nil, err
		}
		response.Refund = &dto.RefundResponse{
			ID:        refund.ID,
			InvoiceID: refund.InvoiceID,
			Amount:    refund.Amount,
			Reason:    refund.Reason,
			CreatedAt: refund.CreatedAt.Format("2006-01-02 15:04"),
			CreatedByUser: &dto.UserSummaryResponse{
				ID:       refund.CreatedByUser.ID,
				Username: refund.CreatedByUser.Username,
			},
		}
	}

	return response, nil
}
