package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrCustomerNotFound          = errors.New("kupac nije pronađen")
	ErrCustomerInactive          = errors.New("kupac nije aktivan; aktivirajte kupca prije kreiranja računa")
	ErrCustomerHasOpenInvoices   = errors.New("kupac ima neizmirene obaveze i ne može biti deaktiviran")
	ErrCustomerHasHistory        = errors.New("kupac ima istoriju računa ili uplata i ne može biti obrisan; deaktivirajte kupca kada izmiri obaveze")
)

func CreateCustomer(customer *models.Customer) error {
	customer.IsActive = true
	return database.DB.Create(customer).Error
}

type CustomerListQuery struct {
	Page            int
	Limit           int
	Search          string
	IncludeInactive bool
}

func buildCustomerListQuery(q CustomerListQuery) *gorm.DB {
	query := database.DB.Model(&models.Customer{})

	if !q.IncludeInactive {
		query = query.Where("is_active = ?", true)
	}

	if q.Search != "" {
		search := strings.ToLower(strings.TrimSpace(q.Search))
		pattern := "%" + search + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(phone) LIKE ?", pattern, pattern)
	}

	return query
}

func GetAllCustomers(q CustomerListQuery) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	if err := buildCustomerListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	err := buildCustomerListQuery(q).
		Order("created_at DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func GetCustomerByID(id uint) (*models.Customer, error) {
	var customer models.Customer
	err := database.DB.Preload("Invoices").First(&customer, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

func UpdateCustomer(customer *models.Customer) error {
	return database.DB.Model(customer).Select("Name", "Phone").Updates(customer).Error
}

func customerHasOpenInvoices(customerID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&models.Invoice{}).
		Where("customer_id = ? AND status IN ?", customerID, []models.InvoiceStatus{
			models.InvoiceStatusUnpaid,
			models.InvoiceStatusPartiallyPaid,
		}).
		Count(&count).Error
	return count > 0, err
}

func UpdateCustomerStatus(id uint, isActive bool) error {
	if !isActive {
		hasOpen, err := customerHasOpenInvoices(id)
		if err != nil {
			return err
		}
		if hasOpen {
			return ErrCustomerHasOpenInvoices
		}
	}

	result := database.DB.Model(&models.Customer{}).Where("id = ?", id).Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

func DeleteCustomer(id uint) error {
	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomerNotFound
		}
		return err
	}

	var invoiceCount int64
	if err := database.DB.Model(&models.Invoice{}).Where("customer_id = ?", id).Count(&invoiceCount).Error; err != nil {
		return err
	}

	var paymentCount int64
	if err := database.DB.Model(&models.Payment{}).Where("customer_id = ?", id).Count(&paymentCount).Error; err != nil {
		return err
	}

	if invoiceCount > 0 || paymentCount > 0 {
		return ErrCustomerHasHistory
	}

	return database.DB.Delete(&customer).Error
}

func validateCustomerForInvoiceTx(tx *gorm.DB, customerID uint) error {
	var customer models.Customer
	if err := tx.First(&customer, customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomerNotFound
		}
		return err
	}
	if !customer.IsActive {
		return ErrCustomerInactive
	}
	return nil
}

func ValidateCustomerForInvoice(customerID uint) error {
	return validateCustomerForInvoiceTx(database.DB, customerID)
}

func GetCustomerFinancialSummary(customerID uint) (*dto.CustomerFinancialSummaryResponse, error) {
	var customer models.Customer
	err := database.DB.First(&customer, customerID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}

	var openInvoicesCount int64

	err = database.DB.Model(&models.Invoice{}).Where("customer_id = ? AND status IN ?", customerID,
		[]models.InvoiceStatus{models.InvoiceStatusUnpaid, models.InvoiceStatusPartiallyPaid}).Count(&openInvoicesCount).Error

	if err != nil {
		return nil, err
	}

	var paymentsCount int64
	err = database.DB.Model(&models.Payment{}).Where("customer_id = ?", customerID).Count(&paymentsCount).Error

	if err != nil {
		return nil, err
	}

	response := dto.CustomerFinancialSummaryResponse{
		ID:                customer.ID,
		Name:              customer.Name,
		Phone:             customer.Phone,
		TotalDebt:         customer.TotalDebt,
		OpenInvoicesCount: openInvoicesCount,
		PaymentsCount:     paymentsCount,
	}

	return &response, nil
}
