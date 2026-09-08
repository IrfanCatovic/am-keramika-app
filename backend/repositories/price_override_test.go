package repositories

import (
	"errors"
	"math"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"
)

func TestCreateInvoiceWithoutPriceOverrideUnchanged(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: p1.ID, Quantity: 2}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	item := invoice.Items[0]
	if item.UnitPrice != 100 || item.TotalPrice != 200 {
		t.Fatalf("unexpected pricing: %+v", item)
	}
	if item.PriceOverridden {
		t.Fatal("expected PriceOverridden=false")
	}
	if item.OriginalUnitPrice == nil || *item.OriginalUnitPrice != 100 {
		t.Fatalf("originalUnitPrice want 100 got %v", item.OriginalUnitPrice)
	}
}

func TestCreateInvoiceStandardPriceOverride(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, _ := seedInvoiceCreateFixtures(t)

	cat := models.Category{}
	database.DB.First(&cat)
	product := models.Product{
		Name: "Std", Slug: "std-override", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 5000, StockQuantity: 10, IsActive: true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{{
			ProductID:     product.ID,
			Quantity:      3,
			PriceOverride: floatPtr(4500),
		}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	item := invoice.Items[0]
	if item.Quantity != 3 {
		t.Fatalf("quantity want 3 got %v", item.Quantity)
	}
	if item.OriginalUnitPrice == nil || *item.OriginalUnitPrice != 5000 {
		t.Fatalf("originalUnitPrice want 5000 got %v", item.OriginalUnitPrice)
	}
	if item.UnitPrice != 4500 || !item.PriceOverridden {
		t.Fatalf("unitPrice/override want 4500/true got %v/%v", item.UnitPrice, item.PriceOverridden)
	}
	if item.TotalPrice != 13500 {
		t.Fatalf("totalPrice want 13500 got %v", item.TotalPrice)
	}
	if invoice.TotalAmount != 13500 {
		t.Fatalf("invoice total want 13500 got %v", invoice.TotalAmount)
	}

	var reloaded models.Product
	database.DB.First(&reloaded, product.ID)
	if reloaded.SalePrice != 5000 {
		t.Fatalf("Product.SalePrice must stay 5000 got %v", reloaded.SalePrice)
	}
	if math.Abs(reloaded.StockQuantity-7) > 1e-9 {
		t.Fatalf("stock want 7 got %v", reloaded.StockQuantity)
	}
}

func TestCreateInvoiceSaleEffectivePlusOverride(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, _ := seedInvoiceCreateFixtures(t)

	cat := models.Category{}
	database.DB.First(&cat)
	// SalePrice 2000, 10% off → RoundUpToTen(1800)=1800
	product := models.Product{
		Name: "Akcija", Slug: "sale-override", CategoryID: cat.ID, Unit: "m2",
		SalePrice: 2000, StockQuantity: 20, IsActive: true,
		IsOnSale: true, DiscountPercent: 10,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{{
			ProductID:     product.ID,
			Quantity:      1,
			PriceOverride: floatPtr(1600),
		}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	item := invoice.Items[0]
	if item.OriginalUnitPrice == nil || *item.OriginalUnitPrice != 1800 {
		t.Fatalf("original want 1800 got %v", item.OriginalUnitPrice)
	}
	if item.UnitPrice != 1600 || !item.PriceOverridden || item.TotalPrice != 1600 {
		t.Fatalf("unexpected item: %+v", item)
	}

	var refreshed models.Product
	database.DB.First(&refreshed, product.ID)
	if refreshed.SalePrice != 2000 || !refreshed.IsOnSale || refreshed.DiscountPercent != 10 {
		t.Fatalf("product catalog fields must stay unchanged: %+v", refreshed)
	}
}

func TestCreateInvoiceRejectsInvalidPriceOverride(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	for _, override := range []float64{0, -10, math.NaN(), math.Inf(1)} {
		v := override
		_, err := CreateInvoice(dto.CreateInvoiceRequest{
			CustomerID: &customer.ID,
			Items: []dto.CreateInvoiceItemRequest{{
				ProductID:     p1.ID,
				Quantity:      1,
				PriceOverride: &v,
			}},
		}, user.ID)
		if !errors.Is(err, pricing.ErrInvalidPriceOverride) {
			t.Fatalf("override=%v want ErrInvalidPriceOverride got %v", override, err)
		}
	}

	var stock models.Product
	database.DB.First(&stock, p1.ID)
	if stock.StockQuantity != 20 {
		t.Fatalf("rejected override must not change stock, got %v", stock.StockQuantity)
	}
}

func TestCreateInvoicePackageWithPriceOverride(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, _, _ := seedInvoiceCreateFixtures(t)

	cat := models.Category{}
	database.DB.First(&cat)
	product := models.Product{
		Name: "Paket", Slug: "pkg-override", CategoryID: cat.ID, Unit: "m2",
		SalePrice: 2000, StockQuantity: 50, IsActive: true,
		SaleByPackage: true, PackageQuantity: 1.44,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{{
			ProductID:     product.ID,
			Quantity:      10,
			PriceOverride: floatPtr(1600),
		}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	item := invoice.Items[0]
	if item.PackageCount != 7 {
		t.Fatalf("packageCount want 7 got %d", item.PackageCount)
	}
	if math.Abs(item.Quantity-10.08) > 1e-9 {
		t.Fatalf("actualQuantity want 10.08 got %v", item.Quantity)
	}
	if item.RequestedQuantity != 10 || item.PackageQuantity != 1.44 || !item.SaleByPackage {
		t.Fatalf("package snapshot unexpected: %+v", item)
	}
	if item.UnitPrice != 1600 || !item.PriceOverridden {
		t.Fatalf("unitPrice want 1600 overridden=true got %v/%v", item.UnitPrice, item.PriceOverridden)
	}
	if item.OriginalUnitPrice == nil || *item.OriginalUnitPrice != 2000 {
		t.Fatalf("originalUnitPrice want 2000 got %v", item.OriginalUnitPrice)
	}
	packagePrice := pricing.CalculatePackagePrice(item.UnitPrice, item.PackageQuantity)
	if packagePrice != 2304 {
		t.Fatalf("packagePrice want 2304 got %v", packagePrice)
	}
	if item.TotalPrice != 16128 || invoice.TotalAmount != 16128 {
		t.Fatalf("line/total want 16128 got item=%v invoice=%v", item.TotalPrice, invoice.TotalAmount)
	}

	var reloaded models.Product
	database.DB.First(&reloaded, product.ID)
	if math.Abs(reloaded.StockQuantity-39.92) > 1e-9 {
		t.Fatalf("stock want 39.92 (50-10.08) got %v", reloaded.StockQuantity)
	}
	var movement models.InventoryMovement
	database.DB.Where("product_id = ? AND movement_type = ?", product.ID, "sale").First(&movement)
	if math.Abs(movement.Quantity-10.08) > 1e-9 {
		t.Fatalf("movement quantity want 10.08 got %v", movement.Quantity)
	}
	if reloaded.SalePrice != 2000 {
		t.Fatalf("SalePrice must stay 2000 got %v", reloaded.SalePrice)
	}
}

func TestCreateInvoicePriceOverrideUpdatesCustomerDebt(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	_, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{{
			ProductID:     p1.ID,
			Quantity:      2,
			PriceOverride: floatPtr(80),
		}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var reloaded models.Customer
	database.DB.First(&reloaded, customer.ID)
	if reloaded.TotalDebt != 160 {
		t.Fatalf("debt want 160 (2*80) got %v", reloaded.TotalDebt)
	}
}

func TestCancelInvoiceUsesSnapshotNotCurrentProductPrice(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	if err := database.DB.AutoMigrate(&models.InvoiceCancellation{}, &models.Refund{}); err != nil {
		t.Fatalf("migrate cancel models: %v", err)
	}
	user, customer, p1, _ := seedInvoiceCreateFixtures(t)

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items: []dto.CreateInvoiceItemRequest{{
			ProductID:     p1.ID,
			Quantity:      2,
			PriceOverride: floatPtr(70),
		}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if invoice.TotalAmount != 140 {
		t.Fatalf("total want 140 got %v", invoice.TotalAmount)
	}

	if err := database.DB.Model(&p1).Update("sale_price", 9999).Error; err != nil {
		t.Fatalf("mutate product price: %v", err)
	}

	cancel, err := CancelInvoice(invoice.ID, dto.CancelInvoiceRequest{Reason: "Greška u ceni"}, user.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancel.DebtReducedAmount != 140 {
		t.Fatalf("debtReduced want 140 (snapshot total) got %v", cancel.DebtReducedAmount)
	}

	var item models.InvoiceItem
	database.DB.First(&item, invoice.Items[0].ID)
	if item.UnitPrice != 70 || item.TotalPrice != 140 {
		t.Fatalf("cancel must keep item snapshot prices: %+v", item)
	}

	var refreshed models.Product
	database.DB.First(&refreshed, p1.ID)
	if refreshed.SalePrice != 9999 {
		t.Fatalf("cancel must not rewrite product sale price, got %v", refreshed.SalePrice)
	}
	if refreshed.StockQuantity != 20 {
		t.Fatalf("stock should be restored to 20, got %v", refreshed.StockQuantity)
	}

	var customerAfter models.Customer
	database.DB.First(&customerAfter, customer.ID)
	if customerAfter.TotalDebt != 0 {
		t.Fatalf("debt after cancel want 0 got %v", customerAfter.TotalDebt)
	}
}
