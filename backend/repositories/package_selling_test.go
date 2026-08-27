package repositories

import (
	"math"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
)

func TestCreateInvoiceEnforcesPackageQuantityAcrossSale(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, product, _ := seedInvoiceCreateFixtures(t)
	product.SaleByPackage = true
	product.PackageQuantity = 2.70
	product.StockQuantity = 27
	if err := database.DB.Save(&product).Error; err != nil {
		t.Fatalf("save product: %v", err)
	}

	invoice, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: product.ID, Quantity: 10}},
	}, user.ID)
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if invoice.TotalAmount != 1080 {
		t.Fatalf("total want 1080 got %v", invoice.TotalAmount)
	}
	if len(invoice.Items) != 1 {
		t.Fatalf("items want 1 got %d", len(invoice.Items))
	}
	item := invoice.Items[0]
	if item.RequestedQuantity != 10 || item.PackageCount != 4 ||
		item.PackageQuantity != 2.70 || !item.SaleByPackage ||
		item.Quantity != 10.80 {
		t.Fatalf("unexpected invoice snapshot: %+v", item)
	}

	var reloaded models.Product
	database.DB.First(&reloaded, product.ID)
	if math.Abs(reloaded.StockQuantity-16.20) > 0.000001 {
		t.Fatalf("stock want 16.2 got %v", reloaded.StockQuantity)
	}
	var movement models.InventoryMovement
	database.DB.Where("product_id = ? AND movement_type = ?", product.ID, "sale").
		First(&movement)
	if movement.Quantity != 10.80 {
		t.Fatalf("movement quantity want 10.8 got %v", movement.Quantity)
	}
}

func TestCreateInvoiceRejectsPackageSaleAgainstActualQuantity(t *testing.T) {
	setupCreateInvoiceTestDB(t)
	user, customer, product, _ := seedInvoiceCreateFixtures(t)
	product.SaleByPackage = true
	product.PackageQuantity = 2.70
	product.StockQuantity = 10.70
	if err := database.DB.Save(&product).Error; err != nil {
		t.Fatalf("save product: %v", err)
	}

	if _, err := CreateInvoice(dto.CreateInvoiceRequest{
		CustomerID: &customer.ID,
		Items:      []dto.CreateInvoiceItemRequest{{ProductID: product.ID, Quantity: 10}},
	}, user.ID); err == nil {
		t.Fatal("expected insufficient stock error")
	}
	var reloaded models.Product
	database.DB.First(&reloaded, product.ID)
	if reloaded.StockQuantity != 10.70 {
		t.Fatalf("failed sale changed stock: %v", reloaded.StockQuantity)
	}
}
