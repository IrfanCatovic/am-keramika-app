package handlers

import (
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"
)

func TestPublicOnlineOrderSnapshotsActualPackageQuantity(t *testing.T) {
	setupOnlineOrderTestDB(t)
	r := setupOnlineOrderRouter()
	_, _, _, product, _ := seedPublicCatalog(t)
	product.SaleByPackage = true
	product.PackageQuantity = 2.70
	product.StockQuantity = 10.80
	if err := database.DB.Save(&product).Error; err != nil {
		t.Fatalf("save product: %v", err)
	}

	w := postPublicOrder(t, r, map[string]any{
		"firstName": "Marko",
		"lastName":  "Marković",
		"phone":     "0651234567",
		"city":      "Beograd",
		"address":   "Ulica 1",
		"items": []map[string]any{
			{"productID": product.ID, "quantity": 10},
		},
	})
	if w.Code != 201 {
		t.Fatalf("want 201 got %d: %s", w.Code, w.Body.String())
	}

	var order models.OnlineOrder
	if err := database.DB.Preload("Items").First(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if len(order.Items) != 1 {
		t.Fatalf("items want 1 got %d", len(order.Items))
	}
	item := order.Items[0]
	if item.RequestedQuantity != 10 || item.PackageCount != 4 ||
		item.PackageQuantity != 2.70 || item.Quantity != 10.80 ||
		item.TotalPrice != 21600 {
		t.Fatalf("unexpected package order item: %+v", item)
	}
	if order.TotalAmount != 21600 {
		t.Fatalf("total want 21600 got %v", order.TotalAmount)
	}

	var unchanged models.Product
	database.DB.First(&unchanged, product.ID)
	if unchanged.StockQuantity != 10.80 {
		t.Fatalf("online order must not decrement stock: %v", unchanged.StockQuantity)
	}
}
