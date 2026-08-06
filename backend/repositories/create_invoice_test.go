package repositories

import (
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCreateInvoiceTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Customer{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Payment{},
		&models.PaymentAllocation{},
		&models.InventoryMovement{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedInvoiceCreateFixtures(t *testing.T) (models.User, models.Customer, models.Product, models.Product) {
	t.Helper()
	user := models.User{Username: "sef", PasswordHash: "x", Role: models.RoleBoss, IsActive: true}
	database.DB.Create(&user)

	customer := models.Customer{Name: "Kupac", Phone: "061", IsActive: true, TotalDebt: 0}
	database.DB.Create(&customer)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)

	p1 := models.Product{
		Name: "A", Slug: "a", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 100, StockQuantity: 20, IsActive: true,
	}
	p2 := models.Product{
		Name: "B", Slug: "b", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 50, StockQuantity: 20, IsActive: true,
	}
	database.DB.Create(&p1)
	database.DB.Create(&p2)
	return user, customer, p1, p2
}

func TestCreateInvoiceCustomerSingleItemTotalAmount(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 2}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if invoice.TotalAmount != 200 {
		t.Fatalf("totalAmount want 200 got %v", invoice.TotalAmount)
	}
	if invoice.Status != models.InvoiceStatusUnpaid {
		t.Fatalf("status want unpaid got %s", invoice.Status)
	}
	if invoice.PaidAmount != 0 {
		t.Fatalf("paidAmount want 0 got %v", invoice.PaidAmount)
	}
}

func TestCreateInvoiceCustomerMultipleItemsSumsTotals(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, p2 := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{
			{ProductID: p1.ID, Quantity: 2}, // 200
			{ProductID: p2.ID, Quantity: 3}, // 150
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if invoice.TotalAmount != 350 {
		t.Fatalf("totalAmount want 350 got %v", invoice.TotalAmount)
	}
	if len(invoice.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(invoice.Items))
	}
}

func TestCreateInvoiceCustomerDebtIncreasesByTotal(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, p2 := seedInvoiceCreateFixtures(t)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{
			{ProductID: p1.ID, Quantity: 1}, // 100
			{ProductID: p2.ID, Quantity: 2}, // 100
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 200 {
		t.Fatalf("debt want 200 got %v", reloaded.TotalDebt)
	}
}

func TestCreateInvoiceCashSaleStillPaid(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, _, p1, _ := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: nil,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 2}},
	}, user.ID)
	if err != nil {
		t.Fatalf("cash create: %v", err)
	}
	if invoice.TotalAmount != 200 {
		t.Fatalf("cash totalAmount want 200 got %v", invoice.TotalAmount)
	}
	if invoice.PaidAmount != 200 {
		t.Fatalf("cash paidAmount want 200 got %v", invoice.PaidAmount)
	}
	if invoice.Status != models.InvoiceStatusPaid {
		t.Fatalf("cash status want paid got %s", invoice.Status)
	}
	if invoice.CustomerID != nil {
		t.Fatal("cash invoice must have nil customerID")
	}
}

func TestCreateInvoiceStockDecreasesOncePerQuantity(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 3}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var product models.Product
	database.DB.First(&product, p1.ID)
	if product.StockQuantity != 17 {
		t.Fatalf("stock want 17 got %v", product.StockQuantity)
	}

	var movementCount int64
	database.DB.Model(&models.InventoryMovement{}).Where("product_id = ?", p1.ID).Count(&movementCount)
	// Customer invoice ne kreira inventory movement po stavci (samo cash).
	if movementCount != 0 {
		t.Fatalf("customer invoice should not create sale movements, got %d", movementCount)
	}
}
