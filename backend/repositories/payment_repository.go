package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreatePayment(req dto.CreatePaymentRequest, createdByUserID uint) (models.Payment, error) {
	tx := database.DB.Begin()

	if tx.Error != nil {
		return models.Payment{}, tx.Error
	}

	var customer models.Customer
	if err := tx.First(&customer, req.CustomerID).Error; err != nil {
		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Payment{}, errors.New("kupac ne postoji")
		}
		return models.Payment{}, err
	}
	if !customer.IsActive {
		tx.Rollback()
		return models.Payment{}, errors.New("kupac nije aktivan")
	}

	if req.TotalAmount <= 0 {
		tx.Rollback()
		return models.Payment{}, errors.New("ukupan iznos uplate mora biti pozitivan")
	}
	var allocationTotal float64 = 0.0

	for _, allocationReq := range req.Allocations {
		allocationTotal += allocationReq.Amount
	}

	if math.Abs(allocationTotal-req.TotalAmount) > 0.01 {
		tx.Rollback()
		return models.Payment{}, errors.New("ukupan iznos uplate se ne poklapa sa raspodelom po racunima")
	}

	if len(req.Allocations) == 0 {
		tx.Rollback()
		return models.Payment{}, errors.New("uplata mora imati bar jedan račun")
	}

	seenInvoiceIDs := make(map[uint]bool)
	for _, allocationReq := range req.Allocations {
		if allocationReq.Amount <= 0 {
			tx.Rollback()
			return models.Payment{}, errors.New("iznos alokacije mora biti pozitivan")
		}
		if seenInvoiceIDs[allocationReq.InvoiceID] {
			tx.Rollback()
			return models.Payment{}, errors.New("isti račun ne može biti dodat dva puta u jednoj uplati")
		}

		seenInvoiceIDs[allocationReq.InvoiceID] = true
	}

	totalAmount := 0.0
	invoicesToUpdate := []models.Invoice{}
	allocationsToCreate := []models.PaymentAllocation{}

	for _, allocationReq := range req.Allocations {
		var invoice models.Invoice

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&invoice, allocationReq.InvoiceID).Error; err != nil {
			tx.Rollback()

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.Payment{}, errors.New("racun ne postoji")
			}
			return models.Payment{}, err
		}

		if invoice.CustomerID == nil || *invoice.CustomerID != customer.ID {
			tx.Rollback()
			return models.Payment{}, errors.New("racun ne pripada kupcu")
		}

		if invoice.Status == models.InvoiceStatusPaid {
			tx.Rollback()
			return models.Payment{}, fmt.Errorf("racun %d je vec placen", invoice.ID)
		}
		if invoice.Status == models.InvoiceStatusCancelled {
			tx.Rollback()
			return models.Payment{}, fmt.Errorf("ne moze se izvrsiti uplata na storniran racun %d", invoice.ID)
		}

		remainingAmount := invoice.TotalAmount - invoice.PaidAmount

		if allocationReq.Amount > remainingAmount+0.0001 {
			tx.Rollback()
			return models.Payment{}, errors.New("iznos uplate ne može biti veći od preostalog duga računa")
		}

		invoice.PaidAmount += allocationReq.Amount

		if math.Abs(invoice.PaidAmount-invoice.TotalAmount) < 0.01 {
			invoice.PaidAmount = invoice.TotalAmount
			invoice.Status = models.InvoiceStatusPaid
		} else {
			invoice.Status = models.InvoiceStatusPartiallyPaid
		}

		totalAmount += allocationReq.Amount
		invoicesToUpdate = append(invoicesToUpdate, invoice)
	}

	payment := models.Payment{
		CustomerID:      &customer.ID,
		CreatedByUserID: createdByUserID,
		TotalAmount:     totalAmount,
	}
	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return models.Payment{}, err
	}

	for _, allocationReq := range req.Allocations {
		allocation := models.PaymentAllocation{
			PaymentID: payment.ID,
			InvoiceID: allocationReq.InvoiceID,
			Amount:    allocationReq.Amount,
		}
		allocationsToCreate = append(allocationsToCreate, allocation)
	}

	if err := tx.Create(&allocationsToCreate).Error; err != nil {
		tx.Rollback()
		return models.Payment{}, err
	}

	for _, invoice := range invoicesToUpdate {
		if err := tx.Model(&models.Invoice{}).
			Where("id = ?", invoice.ID).
			Updates(map[string]interface{}{
				"paid_amount": invoice.PaidAmount,
				"status":      invoice.Status,
			}).Error; err != nil {
			tx.Rollback()
			return models.Payment{}, err
		}
	}

	newCustomerDebt := customer.TotalDebt - totalAmount
	if newCustomerDebt < -0.01 {
		tx.Rollback()
		return models.Payment{}, errors.New("kupac ne moze imati negativan dug")
	}
	if newCustomerDebt < 0 {
		newCustomerDebt = 0
	}

	if err := tx.Model(&models.Customer{}).
		Where("id = ?", customer.ID).
		Update("total_debt", newCustomerDebt).Error; err != nil {
		tx.Rollback()
		return models.Payment{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return models.Payment{}, err
	}

	var createdPayment models.Payment
	if err := database.DB.Preload("Customer").
		Preload("CreatedByUser").
		Preload("Allocations").
		Preload("Allocations.Invoice").
		First(&createdPayment, payment.ID).Error; err != nil {
		return payment, nil
	}
	return createdPayment, nil
}

func GetPaymentsByCustomerID(customerID uint) ([]models.Payment, error) {
	var payments []models.Payment
	var customer models.Customer

	err := database.DB.First(&customer, customerID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kupac nije pronađen")
		}
		return nil, err
	}

	err = database.DB.
		Preload("Customer").
		Preload("CreatedByUser").
		Preload("Allocations").
		Preload("Allocations.Invoice").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Order("id DESC").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func GetPaymentByID(paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	err := database.DB.
		Preload("Customer").
		Preload("CreatedByUser").
		Preload("Allocations").
		Preload("Allocations.Invoice").
		First(&payment, paymentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("uplata nije pronađena")
		}

		return nil, err
	}
	return &payment, nil
}

type PaymentListQuery struct {
	Page       int
	Limit      int
	CustomerID string
	FromDate   *time.Time
	ToDate     *time.Time // exclusive end of selected day
}

func buildPaymentListQuery(q PaymentListQuery) *gorm.DB {
	query := database.DB.Model(&models.Payment{})

	if q.CustomerID != "" {
		query = query.Where("customer_id = ?", q.CustomerID)
	}
	if q.FromDate != nil {
		query = query.Where("created_at >= ?", *q.FromDate)
	}
	if q.ToDate != nil {
		query = query.Where("created_at < ?", *q.ToDate)
	}

	return query
}

func GetAllPayments(q PaymentListQuery) ([]models.Payment, int64, error) {
	var payments []models.Payment
	var total int64

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 50 {
		q.Limit = 50
	}

	countQuery := buildPaymentListQuery(q)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	err := buildPaymentListQuery(q).
		Preload("Customer").
		Preload("CreatedByUser").
		Preload("Allocations").
		Preload("Allocations.Invoice").
		Order("created_at DESC").
		Order("id DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&payments).Error
	if err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}
