package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/mailer"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"

	"github.com/gin-gonic/gin"
)

func setupOnlineOrderStaffRouter() *gin.Engine {
	r := setupOnlineOrderRouter()
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	staff := authorized.Group("/")
	staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
	{
		staff.GET("/online-orders/pending-count", GetOnlineOrdersPendingCount)
		staff.GET("/online-orders", GetOnlineOrders)
		staff.GET("/online-orders/:id", GetOnlineOrderByID)
		staff.POST("/online-orders/:id/confirm", ConfirmOnlineOrder)
		staff.DELETE("/online-orders/:id", DeleteOnlineOrder)
		staff.POST("/customers", CreateCustomer)
	}
	return r
}

func createPendingOrderViaAPI(t *testing.T, r *gin.Engine, productID uint, qty float64) dto.PublicOnlineOrderResponse {
	t.Helper()
	body := map[string]any{
		"firstName": "Irfan",
		"lastName":  "Catovic",
		"phone":     "0631234567",
		"city":      "Tutin",
		"address":   "Ulica 1",
		"items":     []map[string]any{{"productID": productID, "quantity": qty}},
	}
	w := postPublicOrder(t, r, body)
	if w.Code != http.StatusCreated {
		t.Fatalf("create order %d %s", w.Code, w.Body.String())
	}
	var resp dto.PublicOnlineOrderResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

func staffAuthHeader(t *testing.T, r *gin.Engine) string {
	t.Helper()
	pricingCreateUser(t, "radnik_ord", models.RoleWorker)
	token := pricingLogin(t, r, "radnik_ord")
	return "Bearer " + token
}

func TestConfirmOnlineOrderFlow(t *testing.T) {
	setupOnlineOrderTestDB(t)
	rec := &mailer.RecordingMailer{}
	SetOrderMailer(rec)
	t.Setenv("ORDER_NOTIFICATION_EMAIL", "office@am.test")
	defer SetOrderMailer(mailer.NoopMailer{})

	r := setupOnlineOrderStaffRouter()
	_, _, _, featured, _ := seedPublicCatalog(t)
	authHeader := staffAuthHeader(t, r)

	t.Run("public order triggers email best effort", func(t *testing.T) {
		rec.Reset()
		_ = createPendingOrderViaAPI(t, r, featured.ID, 1)
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if rec.Count() >= 1 {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		msg, ok := rec.Last()
		if !ok {
			t.Fatalf("expected email")
		}
		if !strings.Contains(msg.Subject, "#") {
			t.Fatalf("subject %s", msg.Subject)
		}
		if strings.Contains(msg.Body, "purchasePrice") || strings.Contains(msg.Body, "stockQuantity") {
			t.Fatalf("sensitive in email")
		}
	})

	t.Run("email failure still keeps order", func(t *testing.T) {
		rec.Reset()
		rec.SetFailErr(errors.New("smtp down"))
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		time.Sleep(50 * time.Millisecond)
		var order models.OnlineOrder
		if err := database.DB.First(&order, resp.ID).Error; err != nil {
			t.Fatalf("order missing after email fail")
		}
		rec.SetFailErr(nil)
	})

	t.Run("confirm with existing customer uses snapshot price", func(t *testing.T) {
		resp := createPendingOrderViaAPI(t, r, featured.ID, 2)
		// Change current product price — confirmation must ignore it.
		database.DB.Model(&featured).Updates(map[string]any{
			"sale_price":       9999,
			"is_on_sale":       false,
			"discount_percent": 0,
		})

		customer := models.Customer{Name: "Postojeci", Phone: "061", IsActive: true}
		database.DB.Create(&customer)

		var beforeStock models.Product
		database.DB.First(&beforeStock, featured.ID)

		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{CustomerID: &customer.ID})
		req := httptest.NewRequest(http.MethodPost, "/online-orders/"+strconv.FormatUint(uint64(resp.ID), 10)+"/confirm", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var conf dto.ConfirmOnlineOrderResponse
		json.Unmarshal(w.Body.Bytes(), &conf)

		var invoice models.Invoice
		database.DB.Preload("Items").First(&invoice, conf.InvoiceID)
		if invoice.Status != models.InvoiceStatusUnpaid || invoice.PaidAmount != 0 {
			t.Fatalf("invoice %+v", invoice)
		}
		if len(invoice.Items) != 1 || invoice.Items[0].UnitPrice != 2000 {
			t.Fatalf("snapshot price want 2000 got %+v", invoice.Items)
		}
		if invoice.TotalAmount != 4000 {
			t.Fatalf("total %v", invoice.TotalAmount)
		}

		var afterStock models.Product
		database.DB.First(&afterStock, featured.ID)
		if afterStock.StockQuantity != beforeStock.StockQuantity-2 {
			t.Fatalf("stock %v -> %v", beforeStock.StockQuantity, afterStock.StockQuantity)
		}

		var moves int64
		database.DB.Model(&models.InventoryMovement{}).Where("product_id = ? AND movement_type = ?", featured.ID, "sale").Count(&moves)
		if moves < 1 {
			t.Fatalf("expected sale movement")
		}

		var cust models.Customer
		database.DB.First(&cust, customer.ID)
		if cust.TotalDebt != 4000 {
			t.Fatalf("debt %v", cust.TotalDebt)
		}

		var order models.OnlineOrder
		database.DB.First(&order, resp.ID)
		if order.Status != models.OnlineOrderStatusConfirmed || order.InvoiceID == nil || order.ConfirmedAt == nil {
			t.Fatalf("order %+v", order)
		}
	})

	t.Run("confirm with new customer", func(t *testing.T) {
		database.DB.Model(&featured).Update("stock_quantity", 20)
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{
			NewCustomer: &dto.ConfirmOnlineOrderNewCustomer{Name: "Novi Kupac", Phone: "062"},
		})
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/online-orders/%d/confirm", resp.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var cust models.Customer
		if err := database.DB.Where("name = ?", "Novi Kupac").First(&cust).Error; err != nil {
			t.Fatalf("customer not created")
		}
	})

	t.Run("insufficient stock rolls back", func(t *testing.T) {
		database.DB.Model(&featured).Update("stock_quantity", 1)
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		// Order asks 1 but we set stock to 0 after order created
		database.DB.Model(&featured).Update("stock_quantity", 0)
		customer := models.Customer{Name: "StockFail", IsActive: true}
		database.DB.Create(&customer)
		var invBefore, custBefore int64
		database.DB.Model(&models.Invoice{}).Count(&invBefore)
		database.DB.Model(&models.Customer{}).Count(&custBefore)

		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{CustomerID: &customer.ID})
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/online-orders/%d/confirm", resp.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var order models.OnlineOrder
		database.DB.First(&order, resp.ID)
		if order.Status != models.OnlineOrderStatusPending {
			t.Fatalf("should stay pending")
		}
		var invAfter int64
		database.DB.Model(&models.Invoice{}).Count(&invAfter)
		if invAfter != invBefore {
			t.Fatalf("invoice created on rollback")
		}
		database.DB.Model(&featured).Update("stock_quantity", 20)
	})

	t.Run("new customer rolls back on stock fail", func(t *testing.T) {
		database.DB.Model(&featured).Update("stock_quantity", 5)
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		database.DB.Model(&featured).Update("stock_quantity", 0)
		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{
			NewCustomer: &dto.ConfirmOnlineOrderNewCustomer{Name: "Rollback Kupac", Phone: "060"},
		})
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/online-orders/%d/confirm", resp.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%d", w.Code)
		}
		var cust models.Customer
		err := database.DB.Where("name = ?", "Rollback Kupac").First(&cust).Error
		if err == nil {
			t.Fatalf("customer should rollback")
		}
		database.DB.Model(&featured).Update("stock_quantity", 20)
	})

	t.Run("duplicate confirm 409", func(t *testing.T) {
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		customer := models.Customer{Name: "Dup", IsActive: true}
		database.DB.Create(&customer)
		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{CustomerID: &customer.ID})
		do := func() *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/online-orders/%d/confirm", resp.ID), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", authHeader)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			return w
		}
		w1 := do()
		if w1.Code != http.StatusOK {
			t.Fatalf("first %d %s", w1.Code, w1.Body.String())
		}
		w2 := do()
		if w2.Code != http.StatusConflict {
			t.Fatalf("second want 409 got %d %s", w2.Code, w2.Body.String())
		}
		var invCount int64
		database.DB.Model(&models.Invoice{}).Where("customer_id = ?", customer.ID).Count(&invCount)
		if invCount != 1 {
			t.Fatalf("want 1 invoice got %d", invCount)
		}
	})

	t.Run("concurrent confirm single invoice", func(t *testing.T) {
		database.DB.Model(&featured).Update("stock_quantity", 50)
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		customer := models.Customer{Name: "Concurrent", IsActive: true}
		database.DB.Create(&customer)
		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{CustomerID: &customer.ID})

		var wg sync.WaitGroup
		codes := make(chan int, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/online-orders/%d/confirm", resp.ID), bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", authHeader)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				codes <- w.Code
			}()
		}
		wg.Wait()
		close(codes)
		ok := 0
		for code := range codes {
			if code == http.StatusOK {
				ok++
			}
		}
		if ok != 1 {
			t.Fatalf("want exactly 1 success, got ok=%d", ok)
		}
		var invCount int64
		database.DB.Model(&models.Invoice{}).Where("customer_id = ?", customer.ID).Count(&invCount)
		if invCount != 1 {
			t.Fatalf("invoices %d", invCount)
		}
	})

	t.Run("delete pending", func(t *testing.T) {
		var before models.Product
		database.DB.First(&before, featured.ID)
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/online-orders/%d", resp.ID), nil)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var after models.Product
		database.DB.First(&after, featured.ID)
		if after.StockQuantity != before.StockQuantity {
			t.Fatalf("delete mutated stock")
		}
	})

	t.Run("delete confirmed blocked", func(t *testing.T) {
		database.DB.Model(&featured).Update("stock_quantity", 20)
		resp := createPendingOrderViaAPI(t, r, featured.ID, 1)
		customer := models.Customer{Name: "NoDelete", IsActive: true}
		database.DB.Create(&customer)
		body, _ := json.Marshal(dto.ConfirmOnlineOrderRequest{CustomerID: &customer.ID})
		creq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/online-orders/%d/confirm", resp.ID), bytes.NewReader(body))
		creq.Header.Set("Content-Type", "application/json")
		creq.Header.Set("Authorization", authHeader)
		cw := httptest.NewRecorder()
		r.ServeHTTP(cw, creq)
		if cw.Code != http.StatusOK {
			t.Fatalf("confirm %d", cw.Code)
		}
		dreq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/online-orders/%d", resp.ID), nil)
		dreq.Header.Set("Authorization", authHeader)
		dw := httptest.NewRecorder()
		r.ServeHTTP(dw, dreq)
		if dw.Code != http.StatusConflict {
			t.Fatalf("want 409 got %d", dw.Code)
		}
	})

	t.Run("pending count", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/online-orders/pending-count", nil)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d", w.Code)
		}
	})
}
