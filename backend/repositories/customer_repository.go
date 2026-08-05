package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"errors"
	"gorm.io/gorm"
)

func CreateCustomer(customer *models.Customer) error {

	result := database.DB.Create(customer)
	return result.Error
}

func GetAllCustomers(page int, limit int) ([]models.Customer, int64, error) {

	var customers []models.Customer
	var total int64

	query := database.DB.Model(&models.Customer{}) //radimo sa customer tabelom
	offset := (page - 1) * limit                   //offset je broj redova koje preskacemo, racuna se strana 3, limit 20
	//(3-1) * 20 = 40, preskacemo 40 redova i pocinjemo sa 41. redom
	err := query.Count(&total).Error //brojimo ukupan broj redova u tabeli
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil

}

func GetCustomerByID(id uint) (*models.Customer, error) {
	var customer models.Customer

	err := database.DB.Preload("Invoices").First(&customer, id).Error

	if err != nil {
		return nil, err
	}
	return &customer, nil
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
		[]models.InvoiceStatus{models.InvoiceStatusUnpaid, models.InvoiceStatusPartiallyPaid}).Count(&openInvoicesCount).Error //countamo broj otvorenih racuna, nisu placeni ili delimicno

	if err != nil {
		return nil, err
	}

	var paymentsCount int64
	err = database.DB.Model(&models.Payment{}).Where("customer_id = ?", customerID).Count(&paymentsCount).Error //countamo broj placanja

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
