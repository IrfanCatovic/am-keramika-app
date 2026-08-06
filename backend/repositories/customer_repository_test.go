package repositories

import (
	"errors"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCustomerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Customer{}, &models.Invoice{}, &models.InvoiceItem{}, &models.Payment{}, &models.PaymentAllocation{}, &models.User{}, &models.Category{}, &models.Product{}, &models.InventoryMovement{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedCustomer(t *testing.T, name, phone string, active bool) models.Customer {
	t.Helper()
	customer := models.Customer{Name: name, Phone: phone}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	if !active {
		if err := database.DB.Model(&customer).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate customer: %v", err)
		}
		customer.IsActive = false
	}
	return customer
}

func TestUpdateCustomerNameAndPhone(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Stari", "061111111", true)

	customer.Name = "Novi"
	customer.Phone = "062222222"
	if err := UpdateCustomer(&customer); err != nil {
		t.Fatalf("update: %v", err)
	}

	reloaded, err := GetCustomerByID(customer.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if reloaded.Name != "Novi" || reloaded.Phone != "062222222" {
		t.Fatalf("unexpected customer %+v", reloaded)
	}
}

func TestPhoneIsNotUnique(t *testing.T) {
	setupCustomerTestDB(t)
	phone := "061111111"
	if err := CreateCustomer(&models.Customer{Name: "Kupac A", Phone: phone}); err != nil {
		t.Fatalf("create A: %v", err)
	}
	if err := CreateCustomer(&models.Customer{Name: "Kupac B", Phone: phone}); err != nil {
		t.Fatalf("create B with same phone: %v", err)
	}
}

func TestGetAllCustomersSearchByName(t *testing.T) {
	setupCustomerTestDB(t)
	seedCustomer(t, "Dallas Shop", "061111111", true)
	seedCustomer(t, "Drugi", "062222222", true)

	customers, total, err := GetAllCustomers(CustomerListQuery{Page: 1, Limit: 20, Search: "  dallas "})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(customers) != 1 || customers[0].Name != "Dallas Shop" {
		t.Fatalf("expected name search match, got total=%d %+v", total, customers)
	}
}

func TestGetAllCustomersSearchByPhone(t *testing.T) {
	setupCustomerTestDB(t)
	seedCustomer(t, "Kupac", "061999888", true)
	seedCustomer(t, "Drugi", "062222222", true)

	customers, total, err := GetAllCustomers(CustomerListQuery{Page: 1, Limit: 20, Search: "61999"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(customers) != 1 || customers[0].Phone != "061999888" {
		t.Fatalf("expected phone search match, got total=%d %+v", total, customers)
	}
}

func TestGetAllCustomersDefaultHidesInactive(t *testing.T) {
	setupCustomerTestDB(t)
	seedCustomer(t, "Aktivan", "061111111", true)
	seedCustomer(t, "Neaktivan", "062222222", false)

	customers, total, err := GetAllCustomers(CustomerListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(customers) != 1 || customers[0].Name != "Aktivan" {
		t.Fatalf("expected only active customer, got total=%d %+v", total, customers)
	}
}

func TestGetAllCustomersIncludeInactive(t *testing.T) {
	setupCustomerTestDB(t)
	seedCustomer(t, "Aktivan", "061111111", true)
	seedCustomer(t, "Neaktivan", "062222222", false)

	customers, total, err := GetAllCustomers(CustomerListQuery{Page: 1, Limit: 20, IncludeInactive: true})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(customers) != 2 {
		t.Fatalf("expected 2 customers, got total=%d len=%d", total, len(customers))
	}
}

func TestDeactivateCustomerWithoutOpenInvoices(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Kupac", "061111111", true)

	if err := UpdateCustomerStatus(customer.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	reloaded, _ := GetCustomerByID(customer.ID)
	if reloaded.IsActive {
		t.Fatal("expected inactive customer")
	}
}

func TestDeactivateCustomerWithUnpaidInvoice(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Kupac", "061111111", true)
	invoice := models.Invoice{
		CustomerID:      &customer.ID,
		CreatedByUserID: 1,
		Status:          models.InvoiceStatusUnpaid,
		TotalAmount:     100,
	}
	database.DB.Create(&invoice)

	err := UpdateCustomerStatus(customer.ID, false)
	if err != ErrCustomerHasOpenInvoices {
		t.Fatalf("expected open invoices error, got %v", err)
	}
}

func TestDeactivateCustomerWithPartiallyPaidInvoice(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Kupac", "061111111", true)
	invoice := models.Invoice{
		CustomerID:      &customer.ID,
		CreatedByUserID: 1,
		Status:          models.InvoiceStatusPartiallyPaid,
		TotalAmount:     100,
		PaidAmount:      40,
	}
	database.DB.Create(&invoice)

	err := UpdateCustomerStatus(customer.ID, false)
	if err != ErrCustomerHasOpenInvoices {
		t.Fatalf("expected open invoices error, got %v", err)
	}
}

func TestReactivateCustomer(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Kupac", "061111111", true)

	if err := UpdateCustomerStatus(customer.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := UpdateCustomerStatus(customer.ID, true); err != nil {
		t.Fatalf("reactivate: %v", err)
	}

	customers, total, err := GetAllCustomers(CustomerListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(customers) != 1 {
		t.Fatalf("expected reactivated customer in active list, got total=%d", total)
	}
}

func TestDeleteEmptyCustomer(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Prazan", "061111111", true)

	if err := DeleteCustomer(customer.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := GetCustomerByID(customer.ID)
	if err != ErrCustomerNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestDeleteCustomerWithInvoiceHistory(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Kupac", "061111111", true)
	invoice := models.Invoice{
		CustomerID:      &customer.ID,
		CreatedByUserID: 1,
		Status:          models.InvoiceStatusPaid,
		TotalAmount:     100,
		PaidAmount:      100,
	}
	database.DB.Create(&invoice)

	err := DeleteCustomer(customer.ID)
	if err != ErrCustomerHasHistory {
		t.Fatalf("expected history error, got %v", err)
	}
}

func TestDeleteCustomerWithPaymentHistory(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Kupac", "061111111", true)
	customerID := customer.ID
	payment := models.Payment{
		CustomerID:      &customerID,
		CreatedByUserID: 1,
		TotalAmount:     50,
	}
	database.DB.Create(&payment)

	err := DeleteCustomer(customer.ID)
	if err != ErrCustomerHasHistory {
		t.Fatalf("expected history error, got %v", err)
	}
}

func TestValidateCustomerForInvoiceRejectsInactive(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Neaktivan", "061111111", false)

	err := ValidateCustomerForInvoice(customer.ID)
	if err != ErrCustomerInactive {
		t.Fatalf("expected inactive error, got %v", err)
	}
}

func TestCreateInvoiceNilCustomerSkipsActiveValidation(t *testing.T) {
	setupCustomerTestDB(t)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Pločica", Slug: "plocica", CategoryID: cat.ID,
		Unit: "kom", SalePrice: 10, StockQuantity: 5, IsActive: true,
	}
	database.DB.Create(&product)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: nil,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: product.ID, Quantity: 1}},
	}, 1)

	if errors.Is(err, ErrCustomerInactive) || errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("cash invoice must not fail customer validation, got %v", err)
	}
}

func TestValidateCustomerForInvoiceAllowsActive(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, "Aktivan", "061111111", true)

	if err := ValidateCustomerForInvoice(customer.ID); err != nil {
		t.Fatalf("expected active customer to pass, got %v", err)
	}
}
