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

	var movements []models.InventoryMovement
	database.DB.Where("product_id = ? AND movement_type = ?", p1.ID, "sale").Find(&movements)
	if len(movements) != 1 {
		t.Fatalf("customer invoice should create exactly 1 sale movement, got %d", len(movements))
	}
	if movements[0].Quantity != 3 {
		t.Fatalf("sale quantity want 3 got %v", movements[0].Quantity)
	}
	if movements[0].CreatedByUserID != user.ID {
		t.Fatalf("sale CreatedByUserID want %d got %d", user.ID, movements[0].CreatedByUserID)
	}
}

func TestCreateInvoiceCashCreatesExactlyOneSaleMovement(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, _, p1, _ := seedInvoiceCreateFixtures(t)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: nil,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 2}},
	}, user.ID)
	if err != nil {
		t.Fatalf("cash create: %v", err)
	}

	var product models.Product
	database.DB.First(&product, p1.ID)
	if product.StockQuantity != 18 {
		t.Fatalf("stock want 18 got %v", product.StockQuantity)
	}

	var movements []models.InventoryMovement
	database.DB.Where("product_id = ? AND movement_type = ?", p1.ID, "sale").Find(&movements)
	if len(movements) != 1 {
		t.Fatalf("cash invoice should create exactly 1 sale movement, got %d", len(movements))
	}
	if movements[0].Quantity != 2 {
		t.Fatalf("sale quantity want 2 got %v", movements[0].Quantity)
	}
}

func TestCreateInvoiceTwoProductsCreateOneSaleEach(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, p2 := seedInvoiceCreateFixtures(t)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{
			{ProductID: p1.ID, Quantity: 2},
			{ProductID: p2.ID, Quantity: 4},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var m1, m2 []models.InventoryMovement
	database.DB.Where("product_id = ? AND movement_type = ?", p1.ID, "sale").Find(&m1)
	database.DB.Where("product_id = ? AND movement_type = ?", p2.ID, "sale").Find(&m2)
	if len(m1) != 1 || m1[0].Quantity != 2 {
		t.Fatalf("p1 sale movements=%d qty=%v", len(m1), safeQty(m1))
	}
	if len(m2) != 1 || m2[0].Quantity != 4 {
		t.Fatalf("p2 sale movements=%d qty=%v", len(m2), safeQty(m2))
	}

	var totalSales int64
	database.DB.Model(&models.InventoryMovement{}).Where("movement_type = ?", "sale").Count(&totalSales)
	if totalSales != 2 {
		t.Fatalf("total sale movements want 2 got %d", totalSales)
	}
}

func TestCreateInvoiceMovementFailureRollsBackStockAndInvoice(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	// Drop inventory_movements so insert fails inside the invoice transaction.
	if err := database.DB.Migrator().DropTable(&models.InventoryMovement{}); err != nil {
		t.Fatalf("drop movements: %v", err)
	}

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 2}},
	}, user.ID)
	if err == nil {
		t.Fatal("expected create to fail when movement insert fails")
	}

	var product models.Product
	database.DB.First(&product, p1.ID)
	if product.StockQuantity != 20 {
		t.Fatalf("stock should remain 20 after rollback, got %v", product.StockQuantity)
	}

	var invoiceCount int64
	database.DB.Model(&models.Invoice{}).Count(&invoiceCount)
	if invoiceCount != 0 {
		t.Fatalf("invoice should be rolled back, count=%d", invoiceCount)
	}

	var itemCount int64
	database.DB.Model(&models.InvoiceItem{}).Count(&itemCount)
	if itemCount != 0 {
		t.Fatalf("invoice items should be rolled back, count=%d", itemCount)
	}

	var customerReloaded models.Customer
	database.DB.First(&customerReloaded, customer.ID)
	if customerReloaded.TotalDebt != 0 {
		t.Fatalf("customer debt should remain 0 after rollback, got %v", customerReloaded.TotalDebt)
	}

	_ = user
}

func TestCreateInvoiceCustomerThenCancelSaleAndReturnBalance(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	if err := database.DB.AutoMigrate(&models.InvoiceCancellation{}, &models.Refund{}); err != nil {
		t.Fatalf("migrate cancel models: %v", err)
	}

	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 5}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var afterSale models.Product
	database.DB.First(&afterSale, p1.ID)
	if afterSale.StockQuantity != 15 {
		t.Fatalf("after sale stock want 15 got %v", afterSale.StockQuantity)
	}

	_, err = CancelInvoice(invoice.ID, dto.CancelInvoiceRequest{Reason: "Greska u kolicini"}, user.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}

	var afterCancel models.Product
	database.DB.First(&afterCancel, p1.ID)
	if afterCancel.StockQuantity != 20 {
		t.Fatalf("after cancel stock want 20 got %v", afterCancel.StockQuantity)
	}

	var sales, returns []models.InventoryMovement
	database.DB.Where("product_id = ? AND movement_type = ?", p1.ID, "sale").Find(&sales)
	database.DB.Where("product_id = ? AND movement_type = ?", p1.ID, "return").Find(&returns)
	if len(sales) != 1 || sales[0].Quantity != 5 {
		t.Fatalf("sale movements=%v", sales)
	}
	if len(returns) != 1 || returns[0].Quantity != 5 {
		t.Fatalf("return movements=%v", returns)
	}

	// Duplicate cancel must not create another return.
	_, err = CancelInvoice(invoice.ID, dto.CancelInvoiceRequest{Reason: "Ponovo"}, user.ID)
	if err == nil {
		t.Fatal("expected duplicate cancel to fail")
	}
	database.DB.Where("product_id = ? AND movement_type = ?", p1.ID, "return").Find(&returns)
	if len(returns) != 1 {
		t.Fatalf("duplicate cancel must not add return, got %d", len(returns))
	}
}

func safeQty(movements []models.InventoryMovement) float64 {
	if len(movements) == 0 {
		return 0
	}
	return movements[0].Quantity
}

func floatPtr(v float64) *float64 { return &v }

func TestCreateInvoiceCustomerUnpaidNoInitialPayment(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, p2 := seedInvoiceCreateFixtures(t) // p2 SalePrice=50

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:  &customer.ID,
		PaymentMode: dto.InvoicePaymentModeUnpaid,
		Items:       []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if invoice.TotalAmount != 50 || invoice.PaidAmount != 0 {
		t.Fatalf("totals want 50/0 got %v/%v", invoice.TotalAmount, invoice.PaidAmount)
	}
	if invoice.Status != models.InvoiceStatusUnpaid {
		t.Fatalf("status want unpaid got %s", invoice.Status)
	}

	var paymentCount int64
	database.DB.Model(&models.Payment{}).Where("customer_id = ?", customer.ID).Count(&paymentCount)
	if paymentCount != 0 {
		t.Fatalf("unpaid must not create payment, got %d", paymentCount)
	}

	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 50 {
		t.Fatalf("debt want 50 got %v", reloaded.TotalDebt)
	}
}

func TestCreateInvoiceCustomerFullPayment(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, p2 := seedInvoiceCreateFixtures(t)

	// Frontend may send a wrong amount — backend must use invoice total.
	wrong := 1.0
	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModeFull,
		InitialPaymentAmount: &wrong,
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if invoice.TotalAmount != 50 || invoice.PaidAmount != 50 {
		t.Fatalf("totals want 50/50 got %v/%v", invoice.TotalAmount, invoice.PaidAmount)
	}
	if invoice.Status != models.InvoiceStatusPaid {
		t.Fatalf("status want paid got %s", invoice.Status)
	}

	var payments []models.Payment
	database.DB.Where("customer_id = ?", customer.ID).Find(&payments)
	if len(payments) != 1 || payments[0].TotalAmount != 50 {
		t.Fatalf("payment want 1x50 got %+v", payments)
	}
	var alloc models.PaymentAllocation
	database.DB.Where("payment_id = ? AND invoice_id = ?", payments[0].ID, invoice.ID).First(&alloc)
	if alloc.Amount != 50 {
		t.Fatalf("allocation want 50 got %v", alloc.Amount)
	}

	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 0 {
		t.Fatalf("debt want 0 got %v", reloaded.TotalDebt)
	}

	var saleCount int64
	database.DB.Model(&models.InventoryMovement{}).
		Where("product_id = ? AND movement_type = ?", p2.ID, "sale").Count(&saleCount)
	if saleCount != 1 {
		t.Fatalf("sale movements want 1 got %d", saleCount)
	}
}

func TestCreateInvoiceCustomerPartialPayment(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, p2 := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(20),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if invoice.TotalAmount != 50 || invoice.PaidAmount != 20 {
		t.Fatalf("totals want 50/20 got %v/%v", invoice.TotalAmount, invoice.PaidAmount)
	}
	if invoice.Status != models.InvoiceStatusPartiallyPaid {
		t.Fatalf("status want partially_paid got %s", invoice.Status)
	}

	var payments []models.Payment
	database.DB.Where("customer_id = ?", customer.ID).Find(&payments)
	if len(payments) != 1 || payments[0].TotalAmount != 20 {
		t.Fatalf("payment want 1x20 got %+v", payments)
	}
	var alloc models.PaymentAllocation
	database.DB.Where("payment_id = ?", payments[0].ID).First(&alloc)
	if alloc.Amount != 20 || alloc.InvoiceID != invoice.ID {
		t.Fatalf("allocation want 20 on invoice %d got %+v", invoice.ID, alloc)
	}

	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 30 {
		t.Fatalf("debt want 30 got %v", reloaded.TotalDebt)
	}

	var saleCount int64
	database.DB.Model(&models.InventoryMovement{}).
		Where("product_id = ? AND movement_type = ?", p2.ID, "sale").Count(&saleCount)
	if saleCount != 1 {
		t.Fatalf("sale movements want 1 got %d", saleCount)
	}
}

func TestCreateInvoiceCustomerPartialRejectsZeroAndFullAndInvalidMode(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, p2 := seedInvoiceCreateFixtures(t)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(0),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err == nil {
		t.Fatal("partial 0 must fail")
	}

	_, err = CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(50),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err == nil {
		t.Fatal("partial == total must fail")
	}

	_, err = CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(60),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err == nil {
		t.Fatal("partial > total must fail")
	}

	_, err = CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:  &customer.ID,
		PaymentMode: "wire",
		Items:       []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err == nil {
		t.Fatal("invalid mode must fail")
	}

	var invoiceCount int64
	database.DB.Model(&models.Invoice{}).Count(&invoiceCount)
	if invoiceCount != 0 {
		t.Fatalf("rejected creates must leave no invoices, got %d", invoiceCount)
	}
	_ = user
}

func TestCreateInvoicePaymentInsertFailureRollsBack(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, p2 := seedInvoiceCreateFixtures(t)

	if err := database.DB.Migrator().DropTable(&models.Payment{}); err != nil {
		t.Fatalf("drop payments: %v", err)
	}

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:  &customer.ID,
		PaymentMode: dto.InvoicePaymentModeFull,
		Items:       []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err == nil {
		t.Fatal("expected payment insert failure")
	}

	var product models.Product
	database.DB.First(&product, p2.ID)
	if product.StockQuantity != 20 {
		t.Fatalf("stock want 20 after rollback got %v", product.StockQuantity)
	}
	var invoiceCount, itemCount, saleCount int64
	database.DB.Model(&models.Invoice{}).Count(&invoiceCount)
	database.DB.Model(&models.InvoiceItem{}).Count(&itemCount)
	database.DB.Model(&models.InventoryMovement{}).Where("movement_type = ?", "sale").Count(&saleCount)
	if invoiceCount != 0 || itemCount != 0 || saleCount != 0 {
		t.Fatalf("rollback incomplete invoices=%d items=%d sales=%d", invoiceCount, itemCount, saleCount)
	}
	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 0 {
		t.Fatalf("debt want 0 after rollback got %v", reloaded.TotalDebt)
	}
}

func TestCreateInvoiceAllocationFailureRollsBack(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, p2 := seedInvoiceCreateFixtures(t)

	if err := database.DB.Migrator().DropTable(&models.PaymentAllocation{}); err != nil {
		t.Fatalf("drop allocations: %v", err)
	}

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(20),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err == nil {
		t.Fatal("expected allocation insert failure")
	}

	var product models.Product
	database.DB.First(&product, p2.ID)
	if product.StockQuantity != 20 {
		t.Fatalf("stock want 20 after rollback got %v", product.StockQuantity)
	}
	var invoiceCount, paymentCount int64
	database.DB.Model(&models.Invoice{}).Count(&invoiceCount)
	database.DB.Model(&models.Payment{}).Count(&paymentCount)
	if invoiceCount != 0 || paymentCount != 0 {
		t.Fatalf("rollback incomplete invoices=%d payments=%d", invoiceCount, paymentCount)
	}
	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 0 {
		t.Fatalf("debt want 0 after rollback got %v", reloaded.TotalDebt)
	}
}

func TestCreateInvoiceCustomerPartialThenCancelRefundsPaidAmount(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	if err := database.DB.AutoMigrate(&models.InvoiceCancellation{}, &models.Refund{}); err != nil {
		t.Fatalf("migrate cancel: %v", err)
	}
	user, customer, _, p2 := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           &customer.ID,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(20),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	cancel, err := CancelInvoice(invoice.ID, dto.CancelInvoiceRequest{Reason: "Greska u kolicini"}, user.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancel.RefundedAmount != 20 {
		t.Fatalf("refunded want 20 got %v", cancel.RefundedAmount)
	}
	if cancel.DebtReducedAmount != 30 {
		t.Fatalf("debtReduced want 30 got %v", cancel.DebtReducedAmount)
	}

	var product models.Product
	database.DB.First(&product, p2.ID)
	if product.StockQuantity != 20 {
		t.Fatalf("stock restored want 20 got %v", product.StockQuantity)
	}
	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 0 {
		t.Fatalf("debt after cancel want 0 got %v", reloaded.TotalDebt)
	}
}

func TestCreateInvoiceCashIgnoresCustomerPaymentMode(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, _, _, p2 := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID:           nil,
		PaymentMode:          dto.InvoicePaymentModePartial,
		InitialPaymentAmount: floatPtr(10),
		Items:                []dto.CreateInvoiceItemRequest{{ProductID: p2.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("cash create: %v", err)
	}
	if invoice.Status != models.InvoiceStatusPaid || invoice.PaidAmount != 50 {
		t.Fatalf("cash must stay fully paid, status=%s paid=%v", invoice.Status, invoice.PaidAmount)
	}
	var paymentCount int64
	database.DB.Model(&models.Payment{}).Count(&paymentCount)
	if paymentCount != 1 {
		t.Fatalf("cash payment count want 1 got %d", paymentCount)
	}
}

func TestCreateInvoiceUsesEffectiveSalePrice(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, _ := seedInvoiceCreateFixtures(t)

	cat := models.Category{}
	database.DB.First(&cat)

	product := models.Product{
		Name: "Sale", Slug: "sale-eff", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 2350, StockQuantity: 10, IsActive: true,
		IsOnSale: true, DiscountPercent: 15,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: product.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(invoice.Items) != 1 {
		t.Fatalf("items=%d", len(invoice.Items))
	}
	if invoice.Items[0].UnitPrice != 2000 {
		t.Fatalf("unitPrice want 2000 got %v", invoice.Items[0].UnitPrice)
	}
	if invoice.TotalAmount != 2000 {
		t.Fatalf("total want 2000 got %v", invoice.TotalAmount)
	}
}

func TestCreateInvoiceSnapshotKeepsSalePriceAfterSaleEnds(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, _ := seedInvoiceCreateFixtures(t)

	cat := models.Category{}
	database.DB.First(&cat)

	product := models.Product{
		Name: "SaleSnap", Slug: "sale-snap", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 2350, StockQuantity: 10, IsActive: true,
		IsOnSale: true, DiscountPercent: 15,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: product.ID, Quantity: 1}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := database.DB.Model(&product).Updates(map[string]interface{}{
		"is_on_sale": false, "discount_percent": 0,
	}).Error; err != nil {
		t.Fatalf("end sale: %v", err)
	}

	var item models.InvoiceItem
	if err := database.DB.First(&item, invoice.Items[0].ID).Error; err != nil {
		t.Fatalf("reload item: %v", err)
	}
	if item.UnitPrice != 2000 {
		t.Fatalf("snapshot unitPrice want 2000 got %v", item.UnitPrice)
	}

	var refreshed models.Product
	database.DB.First(&refreshed, product.ID)
	if refreshed.SalePrice != 2350 {
		t.Fatalf("regular salePrice should stay 2350 got %v", refreshed.SalePrice)
	}
}
