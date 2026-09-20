package repositories

import (
	"errors"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func boolPtr(v bool) *bool { return &v }

func setupSalesReturnTestDB(t *testing.T) {
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
		&models.Refund{},
		&models.SalesReturn{},
		&models.SalesReturnItem{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedSalesReturnUser(t *testing.T) models.User {
	t.Helper()
	user := models.User{Username: "radnik", PasswordHash: "x", Role: models.RoleWorker, IsActive: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	return user
}

func seedSalesReturnProduct(t *testing.T, name, slug, unit string, stock, salePrice float64, active bool) models.Product {
	t.Helper()
	cat := models.Category{Name: "Keramika-" + slug, Slug: "cat-" + slug, IsActive: true}
	if err := database.DB.Create(&cat).Error; err != nil {
		t.Fatalf("category: %v", err)
	}
	p := models.Product{
		Name: name, Slug: slug, CategoryID: cat.ID, Unit: unit,
		SalePrice: salePrice, StockQuantity: stock, IsActive: active,
	}
	if err := database.DB.Create(&p).Error; err != nil {
		t.Fatalf("product: %v", err)
	}
	if !active {
		if err := database.DB.Model(&p).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate: %v", err)
		}
		p.IsActive = false
	}
	return p
}

func countRows[T any](t *testing.T) int64 {
	t.Helper()
	var n int64
	var zero T
	if err := database.DB.Model(&zero).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestCreateSalesReturnStandardNoCashRefund(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "Bojler", "bojler", "kom", 5, 5000, true)

	result, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		CashRefunded: boolPtr(false),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 2, UnitPrice: 5000},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if result.SalesReturn.TotalAmount != 10000 {
		t.Fatalf("total=%v want 10000", result.SalesReturn.TotalAmount)
	}
	if result.SalesReturn.CashRefunded {
		t.Fatal("cashRefunded want false")
	}
	if result.RefundID != nil {
		t.Fatalf("refundID want nil got %v", *result.RefundID)
	}
	if result.SalesReturn.CreatedByUserID != user.ID {
		t.Fatalf("createdBy=%d want %d", result.SalesReturn.CreatedByUserID, user.ID)
	}

	var stored models.Product
	database.DB.First(&stored, product.ID)
	if stored.StockQuantity != 7 {
		t.Fatalf("stock=%v want 7", stored.StockQuantity)
	}

	var movements []models.InventoryMovement
	database.DB.Find(&movements)
	if len(movements) != 1 || movements[0].MovementType != "return" || movements[0].Quantity != 2 {
		t.Fatalf("movements=%+v", movements)
	}
	if movements[0].CreatedByUserID != user.ID {
		t.Fatalf("movement audit=%d want %d", movements[0].CreatedByUserID, user.ID)
	}
	if countRows[models.Refund](t) != 0 {
		t.Fatal("expected 0 refunds")
	}
}

func TestCreateSalesReturnCashRefund(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "Bojler", "bojler", "kom", 5, 5000, true)

	result, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		Description:  "Marko Marković - višak materijala",
		CashRefunded: boolPtr(true),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 2, UnitPrice: 5000},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if result.SalesReturn.TotalAmount != 10000 || !result.SalesReturn.CashRefunded {
		t.Fatalf("return=%+v", result.SalesReturn)
	}
	if result.RefundID == nil {
		t.Fatal("expected refundID")
	}

	var stored models.Product
	database.DB.First(&stored, product.ID)
	if stored.StockQuantity != 7 {
		t.Fatalf("stock=%v want 7", stored.StockQuantity)
	}

	var refund models.Refund
	if err := database.DB.First(&refund, *result.RefundID).Error; err != nil {
		t.Fatalf("refund: %v", err)
	}
	if refund.Amount != 10000 {
		t.Fatalf("refund amount=%v want 10000", refund.Amount)
	}
	if refund.InvoiceID != nil {
		t.Fatalf("invoiceID must stay nil, got %v", *refund.InvoiceID)
	}
	if refund.SalesReturnID == nil || *refund.SalesReturnID != result.SalesReturn.ID {
		t.Fatalf("salesReturnID=%v want %d", refund.SalesReturnID, result.SalesReturn.ID)
	}
	if refund.CreatedByUserID != user.ID {
		t.Fatalf("refund audit=%d want %d", refund.CreatedByUserID, user.ID)
	}
	if countRows[models.InventoryMovement](t) != 1 {
		t.Fatal("expected 1 return movement")
	}
}

func TestCreateSalesReturnPackageQuantityNotCeiled(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	cat := models.Category{Name: "Pločice", Slug: "plocice", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Calacatta", Slug: "calacatta", CategoryID: cat.ID, Unit: "m²",
		SalePrice: 2000, StockQuantity: 5.40, IsActive: true,
		SaleByPackage: true, PackageQuantity: 1.44,
	}
	database.DB.Create(&product)

	result, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		CashRefunded: boolPtr(false),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 20, UnitPrice: 1600},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	item := result.SalesReturn.Items[0]
	if item.Quantity != 20 {
		t.Fatalf("quantity=%v want 20", item.Quantity)
	}
	if item.TotalPrice != 32000 {
		t.Fatalf("totalPrice=%v want 32000", item.TotalPrice)
	}
	if !item.SaleByPackage || item.PackageQuantity != 1.44 {
		t.Fatalf("package snapshot=%+v", item)
	}

	var stored models.Product
	database.DB.First(&stored, product.ID)
	wantStock := pricing.RoundQuantity(5.40 + 20)
	if stored.StockQuantity != wantStock {
		t.Fatalf("stock=%v want %v (not packaged 25.56)", stored.StockQuantity, wantStock)
	}
	if stored.SalePrice != 2000 {
		t.Fatalf("salePrice changed to %v", stored.SalePrice)
	}
}

func TestCreateSalesReturnMultipleItemsWithCashRefund(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	tiles := seedSalesReturnProduct(t, "Calacatta", "calacatta", "m²", 10, 2000, true)
	database.DB.Model(&tiles).Updates(map[string]any{"sale_by_package": true, "package_quantity": 1.44})
	faucet := seedSalesReturnProduct(t, "Slavina X", "slavina-x", "kom", 4, 5000, true)

	customer := models.Customer{Name: "Kupac", Phone: "061", IsActive: true, TotalDebt: 8000}
	database.DB.Create(&customer)
	invoice := models.Invoice{
		CreatedByUserID: user.ID,
		CustomerID:      &customer.ID,
		TotalAmount:     8000,
		PaidAmount:      0,
		Status:          models.InvoiceStatusUnpaid,
	}
	database.DB.Create(&invoice)

	result, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		Description:  "Marko Marković - višak robe",
		CashRefunded: boolPtr(true),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: tiles.ID, Quantity: 20, UnitPrice: 1600},
			{ProductID: faucet.ID, Quantity: 2, UnitPrice: 4500},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if result.SalesReturn.TotalAmount != 41000 {
		t.Fatalf("total=%v want 41000", result.SalesReturn.TotalAmount)
	}
	if len(result.SalesReturn.Items) != 2 {
		t.Fatalf("items=%d want 2", len(result.SalesReturn.Items))
	}
	if result.RefundID == nil {
		t.Fatal("expected one refund")
	}

	var refund models.Refund
	database.DB.First(&refund, *result.RefundID)
	if refund.Amount != 41000 {
		t.Fatalf("refund=%v want 41000", refund.Amount)
	}
	if countRows[models.Refund](t) != 1 {
		t.Fatal("expected exactly 1 refund")
	}
	if countRows[models.InventoryMovement](t) != 2 {
		t.Fatal("expected 2 movements")
	}

	var storedTiles, storedFaucet models.Product
	database.DB.First(&storedTiles, tiles.ID)
	database.DB.First(&storedFaucet, faucet.ID)
	if storedTiles.StockQuantity != 30 {
		t.Fatalf("tiles stock=%v want 30", storedTiles.StockQuantity)
	}
	if storedFaucet.StockQuantity != 6 {
		t.Fatalf("faucet stock=%v want 6", storedFaucet.StockQuantity)
	}

	var storedCustomer models.Customer
	database.DB.First(&storedCustomer, customer.ID)
	if storedCustomer.TotalDebt != 8000 {
		t.Fatalf("customer debt changed: %v", storedCustomer.TotalDebt)
	}
	var storedInvoice models.Invoice
	database.DB.First(&storedInvoice, invoice.ID)
	if storedInvoice.TotalAmount != 8000 || storedInvoice.Status != models.InvoiceStatusUnpaid {
		t.Fatalf("invoice mutated: %+v", storedInvoice)
	}
	if countRows[models.Payment](t) != 0 {
		t.Fatal("payments must not be created")
	}
}

func TestCreateSalesReturnRollbackMissingProduct(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "Bojler", "bojler", "kom", 5, 5000, true)

	_, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		CashRefunded: boolPtr(true),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 2, UnitPrice: 5000},
			{ProductID: 99999, Quantity: 1, UnitPrice: 100},
		},
	}, user.ID)
	if !errors.Is(err, pricing.ErrSalesReturnProductNotFound) {
		t.Fatalf("want product not found, got %v", err)
	}

	var stored models.Product
	database.DB.First(&stored, product.ID)
	if stored.StockQuantity != 5 {
		t.Fatalf("stock changed on rollback: %v", stored.StockQuantity)
	}
	if countRows[models.SalesReturn](t) != 0 {
		t.Fatal("expected 0 sales returns")
	}
	if countRows[models.SalesReturnItem](t) != 0 {
		t.Fatal("expected 0 sales return items")
	}
	if countRows[models.InventoryMovement](t) != 0 {
		t.Fatal("expected 0 movements")
	}
	if countRows[models.Refund](t) != 0 {
		t.Fatal("expected 0 refunds")
	}
}

func TestCreateSalesReturnInactiveProductAllowed(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "Stari bojler", "stari-bojler", "kom", 3, 4000, false)

	result, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		CashRefunded: boolPtr(false),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 1, UnitPrice: 4000},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("inactive product should be returnable: %v", err)
	}

	var stored models.Product
	database.DB.First(&stored, product.ID)
	if stored.StockQuantity != 4 {
		t.Fatalf("stock=%v want 4", stored.StockQuantity)
	}
	if stored.IsActive {
		t.Fatal("product must stay inactive")
	}
	if result.SalesReturn.Items[0].ProductName != "Stari bojler" {
		t.Fatalf("snapshot name=%q", result.SalesReturn.Items[0].ProductName)
	}
}

func TestCreateSalesReturnDoesNotChangeProductPrice(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "Calacatta", "calacatta", "m²", 8, 2000, true)

	_, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		CashRefunded: boolPtr(false),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 20, UnitPrice: 1600},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var stored models.Product
	database.DB.First(&stored, product.ID)
	if stored.SalePrice != 2000 {
		t.Fatalf("salePrice=%v want 2000", stored.SalePrice)
	}
}

func TestCreateSalesReturnRequiresCashRefunded(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "X", "x", "kom", 1, 100, true)

	_, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 1, UnitPrice: 100},
		},
	}, user.ID)
	if !errors.Is(err, pricing.ErrSalesReturnCashRefundedRequired) {
		t.Fatalf("want cashRefunded required, got %v", err)
	}
}

func TestCreateSalesReturnRejectsDuplicateProductID(t *testing.T) {
	setupSalesReturnTestDB(t)
	user := seedSalesReturnUser(t)
	product := seedSalesReturnProduct(t, "X", "x", "kom", 10, 100, true)

	_, err := CreateSalesReturn(dto.CreateSalesReturnRequest{
		CashRefunded: boolPtr(false),
		Items: []dto.CreateSalesReturnItemRequest{
			{ProductID: product.ID, Quantity: 1, UnitPrice: 100},
			{ProductID: product.ID, Quantity: 2, UnitPrice: 100},
		},
	}, user.ID)
	if !errors.Is(err, pricing.ErrSalesReturnDuplicateProduct) {
		t.Fatalf("want duplicate error, got %v", err)
	}
	if countRows[models.SalesReturn](t) != 0 {
		t.Fatal("duplicate must have no side effects")
	}
}
