package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"github.com/gin-gonic/gin"
)

func setupOnlineOrderTestDB(t *testing.T) {
	t.Helper()
	setupPricingHandlerTestDB(t)
	if err := database.DB.AutoMigrate(&models.OnlineOrder{}, &models.OnlineOrderItem{}, &models.InventoryMovement{}, &models.Invoice{}, &models.InvoiceItem{}, &models.Payment{}); err != nil {
		t.Fatalf("migrate online order: %v", err)
	}
}

func setupOnlineOrderRouter() *gin.Engine {
	r := setupSaleRouter()
	r.POST("/public/orders", CreatePublicOnlineOrder)
	return r
}

func postPublicOrder(t *testing.T, r *gin.Engine, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/public/orders", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreatePublicOnlineOrder(t *testing.T) {
	setupOnlineOrderTestDB(t)
	r := setupOnlineOrderRouter()
	_, _, _, featured, regular := seedPublicCatalog(t)

	base := func(overrides map[string]any) map[string]any {
		body := map[string]any{
			"firstName": "Marko",
			"lastName":  "Marković",
			"phone":     "0651234567",
			"city":      "Beograd",
			"address":   "Ulica 1",
			"items": []map[string]any{
				{"productID": featured.ID, "quantity": 2},
			},
		}
		for k, v := range overrides {
			body[k] = v
		}
		return body
	}

	t.Run("valid order pending", func(t *testing.T) {
		w := postPublicOrder(t, r, base(nil))
		if w.Code != http.StatusCreated {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var resp dto.PublicOnlineOrderResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.ID == 0 || resp.Status != "pending" {
			t.Fatalf("%+v", resp)
		}
		// featured: sale 2350 @15% → effective 2000; qty 2 → 4000
		if resp.TotalAmount != 4000 {
			t.Fatalf("total want 4000 got %v", resp.TotalAmount)
		}
		raw := w.Body.String()
		if strings.Contains(raw, "stockQuantity") || strings.Contains(raw, "purchasePrice") {
			t.Fatalf("leaked sensitive: %s", raw)
		}

		var order models.OnlineOrder
		database.DB.Preload("Items").First(&order, resp.ID)
		if len(order.Items) != 1 {
			t.Fatalf("items %d", len(order.Items))
		}
		if order.Items[0].UnitPrice != 2000 || order.Items[0].TotalPrice != 4000 {
			t.Fatalf("snapshot %+v", order.Items[0])
		}
		if order.Items[0].ProductName == "" || order.Items[0].Unit == "" {
			t.Fatalf("missing snapshot name/unit")
		}
	})

	t.Run("required fields", func(t *testing.T) {
		cases := []struct {
			name string
			body map[string]any
		}{
			{"firstName", base(map[string]any{"firstName": ""})},
			{"lastName", base(map[string]any{"lastName": "  "})},
			{"phone", base(map[string]any{"phone": "12"})},
			{"city", base(map[string]any{"city": ""})},
			{"address", base(map[string]any{"address": ""})},
			{"items", base(map[string]any{"items": []any{}})},
		}
		for _, tc := range cases {
			w := postPublicOrder(t, r, tc.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("%s: want 400 got %d %s", tc.name, w.Code, w.Body.String())
			}
		}
	})

	t.Run("optional email ok", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{"email": "marko@example.com", "note": "Pozvati posle 17h"}))
		if w.Code != http.StatusCreated {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{"email": "not-an-email"}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
	})

	t.Run("quantity zero", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": featured.ID, "quantity": 0}},
		}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": featured.ID, "quantity": 999}},
		}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Nema dovoljno") {
			t.Fatalf("%s", w.Body.String())
		}
		if strings.Contains(w.Body.String(), "stockQuantity") {
			t.Fatalf("leaked stock")
		}
	})

	t.Run("zero stock product", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": regular.ID, "quantity": 1}},
		}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
	})

	t.Run("inactive product", func(t *testing.T) {
		var inactive models.Product
		database.DB.Where("slug = ?", "sakriven").First(&inactive)
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": inactive.ID, "quantity": 1}},
		}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "nije dostupan") {
			t.Fatalf("%s", w.Body.String())
		}
	})

	t.Run("inactive category", func(t *testing.T) {
		var orphan models.Product
		database.DB.Where("slug = ?", "u-staroj").First(&orphan)
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": orphan.ID, "quantity": 1}},
		}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
	})

	t.Run("client fake price ignored", func(t *testing.T) {
		body := base(nil)
		body["totalAmount"] = 1
		body["items"] = []map[string]any{
			{
				"productID":  featured.ID,
				"quantity":   1,
				"unitPrice":  1,
				"totalPrice": 1,
				"name":       "Hacked",
			},
		}
		w := postPublicOrder(t, r, body)
		if w.Code != http.StatusCreated {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var resp dto.PublicOnlineOrderResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.TotalAmount != 2000 {
			t.Fatalf("want backend price 2000 got %v", resp.TotalAmount)
		}
		var item models.OnlineOrderItem
		database.DB.Where("online_order_id = ?", resp.ID).First(&item)
		if item.ProductName == "Hacked" || item.UnitPrice != 2000 {
			t.Fatalf("client snapshot leaked: %+v", item)
		}
	})

	t.Run("stock not mutated and no side effects", func(t *testing.T) {
		var before models.Product
		database.DB.First(&before, featured.ID)
		var invBefore, payBefore, moveBefore int64
		database.DB.Model(&models.Invoice{}).Count(&invBefore)
		database.DB.Model(&models.Payment{}).Count(&payBefore)
		database.DB.Model(&models.InventoryMovement{}).Count(&moveBefore)

		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": featured.ID, "quantity": 1}},
		}))
		if w.Code != http.StatusCreated {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}

		var after models.Product
		database.DB.First(&after, featured.ID)
		if after.StockQuantity != before.StockQuantity {
			t.Fatalf("stock mutated %v -> %v", before.StockQuantity, after.StockQuantity)
		}
		var invAfter, payAfter, moveAfter int64
		database.DB.Model(&models.Invoice{}).Count(&invAfter)
		database.DB.Model(&models.Payment{}).Count(&payAfter)
		database.DB.Model(&models.InventoryMovement{}).Count(&moveAfter)
		if invAfter != invBefore || payAfter != payBefore || moveAfter != moveBefore {
			t.Fatalf("side effects created")
		}
	})

	t.Run("multiple items total", func(t *testing.T) {
		// Restock regular for this test
		database.DB.Model(&regular).Update("stock_quantity", 5)
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{
				{"productID": featured.ID, "quantity": 1},
				{"productID": regular.ID, "quantity": 2},
			},
		}))
		if w.Code != http.StatusCreated {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var resp dto.PublicOnlineOrderResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		// 2000 + 2*1800 = 5600
		if resp.TotalAmount != 5600 {
			t.Fatalf("want 5600 got %v", resp.TotalAmount)
		}
	})

	t.Run("honeypot rejected", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{"website": "http://spam.test"}))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
	})

	t.Run("sale snapshot stays after product price change", func(t *testing.T) {
		w := postPublicOrder(t, r, base(map[string]any{
			"items": []map[string]any{{"productID": featured.ID, "quantity": 1}},
		}))
		var resp dto.PublicOnlineOrderResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		database.DB.Model(&featured).Updates(map[string]any{
			"is_on_sale":        false,
			"discount_percent":  0,
			"sale_price":        9999,
		})
		var item models.OnlineOrderItem
		database.DB.Where("online_order_id = ?", resp.ID).First(&item)
		if item.UnitPrice != 2000 {
			t.Fatalf("snapshot changed to %v", item.UnitPrice)
		}
	})
}
