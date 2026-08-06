package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/storage"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var smokeJPEG = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

func setupMVPSmokeDB(t *testing.T) {
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
		&models.ProductGroup{},
		&models.Product{},
		&models.ProductImage{},
		&models.InventoryMovement{},
		&models.Customer{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Payment{},
		&models.PaymentAllocation{},
		&models.InvoiceCancellation{},
		&models.Refund{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-mvp-smoke")
	handlers.SetImageStorage(storage.NewFakeStorage())
}

func setupMVPSmokeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		authorized.GET("/auth/me", handlers.GetMe)

		bossOnly := authorized.Group("/")
		bossOnly.Use(middleware.RequireRoles(models.RoleBoss))
		{
			bossOnly.POST("/users", handlers.CreateUser)
		}

		reports := authorized.Group("/reports")
		reports.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager))
		{
			reports.GET("/daily", handlers.GetDailyReport)
		}

		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/categories", handlers.CreateCategory)
			staff.POST("/product-groups", handlers.CreateProductGroup)
			staff.POST("/products", handlers.CreateProduct)
			staff.GET("/products/:id", handlers.GetProductById)
			staff.POST("/products/:id/images", handlers.UploadProductImages)
			staff.POST("/customers", handlers.CreateCustomer)
			staff.GET("/customers/:id", handlers.GetCustomerByID)
			staff.GET("/customers/:id/financial-summary", handlers.GetCustomerFinancialSummary)
			staff.POST("/invoices", handlers.CreateInvoice)
			staff.GET("/invoices", handlers.GetAllInvoices)
			staff.GET("/invoices/:id", handlers.GetInvoiceByID)
			staff.PUT("/invoices/:id/cancel", handlers.CancelInvoice)
			staff.POST("/payments", handlers.CreatePayment)
			staff.GET("/inventory/low-stock", handlers.GetLowStock)
		}
	}
	return r
}

func smokeLogin(t *testing.T, r *gin.Engine, username, password string) (string, uint) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login %s: status %d body %s", username, w.Code, w.Body.String())
	}
	var resp dto.LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("login decode: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("missing token")
	}
	return resp.Token, resp.User.ID
}

func smokeJSON(t *testing.T, r *gin.Engine, method, path, token string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var body *bytes.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		body = bytes.NewReader(raw)
	} else {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestMVPSmokeHappyPath(t *testing.T) {
	setupMVPSmokeDB(t)
	r := setupMVPSmokeRouter()

	// 1. Pripremi aktivnog sefa
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	boss := models.User{
		Username:     "sef",
		PasswordHash: hash,
		Role:         models.RoleBoss,
		IsActive:     true,
	}
	if err := repositories.CreateUser(&boss); err != nil {
		t.Fatalf("create boss: %v", err)
	}

	// 2. Login
	bossToken, bossID := smokeLogin(t, r, "sef", "password123")
	if bossID != boss.ID {
		t.Fatalf("expected boss id %d, got %d", boss.ID, bossID)
	}

	// 3. GET /auth/me
	w := smokeJSON(t, r, http.MethodGet, "/auth/me", bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("auth/me: %d %s", w.Code, w.Body.String())
	}
	var me dto.AuthUserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &me); err != nil {
		t.Fatalf("auth/me decode: %v", err)
	}
	if me.Username != "sef" || me.Role != models.RoleBoss {
		t.Fatalf("unexpected /auth/me: %+v", me)
	}

	// 4. Kreiraj kategoriju
	w = smokeJSON(t, r, http.MethodPost, "/categories", bossToken, map[string]string{"name": "Keramika"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: %d %s", w.Code, w.Body.String())
	}
	var catResp struct {
		Data dto.CategoryResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &catResp)
	if catResp.Data.ID == 0 {
		t.Fatal("missing category id")
	}
	categoryID := catResp.Data.ID

	// 5. Kreiraj grupu
	w = smokeJSON(t, r, http.MethodPost, "/product-groups", bossToken, map[string]interface{}{
		"name":       "Verona",
		"categoryID": categoryID,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create group: %d %s", w.Code, w.Body.String())
	}
	var groupResp struct {
		Data dto.ProductGroupResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &groupResp)
	groupID := groupResp.Data.ID
	if groupID == 0 {
		t.Fatal("missing group id")
	}

	// 6. Kreiraj proizvod (salePrice + stockQuantity via HTTP)
	// minStockQuantity nije dio CreateProductRequest API contracta — postavlja se direktno u DB nakon create.
	initialStock := 10.0
	minStock := 8.0
	salePrice := 100.0
	w = smokeJSON(t, r, http.MethodPost, "/products", bossToken, map[string]interface{}{
		"name":          "Verona Beige",
		"categoryID":    categoryID,
		"groupID":       groupID,
		"unit":          "m2",
		"salePrice":     salePrice,
		"stockQuantity": initialStock,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", w.Code, w.Body.String())
	}
	var product dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &product)
	if product.ID == 0 {
		t.Fatal("missing product id")
	}
	if err := database.DB.Model(&models.Product{}).Where("id = ?", product.ID).
		Update("min_stock_quantity", minStock).Error; err != nil {
		t.Fatalf("set minStockQuantity via DB (API field missing on create): %v", err)
	}

	// 7. Upload slike (fake storage) + primary
	var imgBody bytes.Buffer
	writer := multipart.NewWriter(&imgBody)
	part, _ := writer.CreateFormFile("images", "test.jpg")
	part.Write(smokeJPEG)
	writer.Close()
	imgReq := httptest.NewRequest(http.MethodPost, "/products/"+strconv.FormatUint(uint64(product.ID), 10)+"/images", &imgBody)
	imgReq.Header.Set("Authorization", "Bearer "+bossToken)
	imgReq.Header.Set("Content-Type", writer.FormDataContentType())
	imgW := httptest.NewRecorder()
	r.ServeHTTP(imgW, imgReq)
	if imgW.Code != http.StatusCreated {
		t.Fatalf("upload image: %d %s", imgW.Code, imgW.Body.String())
	}
	w = smokeJSON(t, r, http.MethodGet, "/products/"+strconv.FormatUint(uint64(product.ID), 10), bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("product detail: %d %s", w.Code, w.Body.String())
	}
	var productDetail dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &productDetail)
	if len(productDetail.Images) == 0 || !productDetail.Images[0].IsPrimary {
		t.Fatalf("expected primary image after upload, got %+v", productDetail.Images)
	}

	// 8. Kreiraj kupca
	w = smokeJSON(t, r, http.MethodPost, "/customers", bossToken, map[string]string{
		"name":  "Dallas Shop",
		"phone": "061111111",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create customer: %d %s", w.Code, w.Body.String())
	}
	var customerWrap struct {
		Customer dto.CustomerResponse `json:"customer"`
	}
	json.Unmarshal(w.Body.Bytes(), &customerWrap)
	customerID := customerWrap.Customer.ID
	if customerID == 0 {
		t.Fatal("missing customer id")
	}

	// 9. Kreiraj račun za kupca
	qty := 3.0
	w = smokeJSON(t, r, http.MethodPost, "/invoices", bossToken, map[string]interface{}{
		"customerID": customerID,
		"items": []map[string]interface{}{
			{"productID": product.ID, "quantity": qty},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create invoice: %d %s", w.Code, w.Body.String())
	}
	var invoiceWrap struct {
		Invoice dto.InvoiceResponse `json:"invoice"`
	}
	json.Unmarshal(w.Body.Bytes(), &invoiceWrap)
	invoice := invoiceWrap.Invoice
	if invoice.ID == 0 {
		t.Fatal("missing invoice id")
	}

	expectedTotal := salePrice * qty
	var dbInvoice models.Invoice
	if err := database.DB.First(&dbInvoice, invoice.ID).Error; err != nil {
		t.Fatalf("reload invoice: %v", err)
	}
	if dbInvoice.CreatedByUserID != bossID {
		t.Fatalf("CreatedByUserID want %d got %d", bossID, dbInvoice.CreatedByUserID)
	}

	var dbProduct models.Product
	database.DB.First(&dbProduct, product.ID)
	if dbProduct.StockQuantity != initialStock-qty {
		t.Fatalf("stock after invoice: want %v got %v", initialStock-qty, dbProduct.StockQuantity)
	}

	var dbCustomer models.Customer
	database.DB.First(&dbCustomer, customerID)

	// Dokumentovana greška u CreateInvoice: totalAmount se ne uvećava za kupca
	// (totalAmount += totalPrice je unutar if customerID == nil).
	if dbInvoice.TotalAmount != expectedTotal {
		t.Fatalf("FOUND BUG CreateInvoice: TotalAmount want %v got %v (customer debt=%v, status=%s). "+
			"totalAmount += totalPrice je unutar cash-only bloka; dug i total ostaju 0 za račun sa kupcem",
			expectedTotal, dbInvoice.TotalAmount, dbCustomer.TotalDebt, dbInvoice.Status)
	}
	if dbInvoice.Status != models.InvoiceStatusUnpaid {
		t.Fatalf("invoice status want unpaid got %s", dbInvoice.Status)
	}
	if dbCustomer.TotalDebt != expectedTotal {
		t.Fatalf("customer debt want %v got %v", expectedTotal, dbCustomer.TotalDebt)
	}
	if invoice.Status != string(models.InvoiceStatusUnpaid) {
		t.Fatalf("response status want unpaid got %s", invoice.Status)
	}

	// 10. Djelimična uplata
	partial := 150.0
	w = smokeJSON(t, r, http.MethodPost, "/payments", bossToken, map[string]interface{}{
		"customerID":  customerID,
		"totalAmount": partial,
		"allocations": []map[string]interface{}{
			{"invoiceID": invoice.ID, "amount": partial},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create payment: %d %s", w.Code, w.Body.String())
	}

	w = smokeJSON(t, r, http.MethodGet, "/invoices/"+strconv.FormatUint(uint64(invoice.ID), 10), bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get invoice: %d %s", w.Code, w.Body.String())
	}
	var afterPay dto.InvoiceResponse
	json.Unmarshal(w.Body.Bytes(), &afterPay)
	if afterPay.PaidAmount != partial {
		t.Fatalf("paidAmount want %v got %v", partial, afterPay.PaidAmount)
	}
	if afterPay.RemainingAmount != expectedTotal-partial {
		t.Fatalf("remainingAmount want %v got %v", expectedTotal-partial, afterPay.RemainingAmount)
	}
	if afterPay.Status != string(models.InvoiceStatusPartiallyPaid) {
		t.Fatalf("status want partially_paid got %s", afterPay.Status)
	}
	database.DB.First(&dbCustomer, customerID)
	if dbCustomer.TotalDebt != expectedTotal-partial {
		t.Fatalf("debt after payment want %v got %v", expectedTotal-partial, dbCustomer.TotalDebt)
	}

	// 11. GET /invoices filteri
	today := time.Now().In(mustBelgrade(t)).Format("2006-01-02")
	filterPath := fmt.Sprintf(
		"/invoices?page=1&limit=20&status=partially_paid&customerID=%d&fromDate=%s&toDate=%s&search=dallas",
		customerID, today, today,
	)
	w = smokeJSON(t, r, http.MethodGet, filterPath, bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list invoices: %d %s", w.Code, w.Body.String())
	}
	var list dto.PaginatedInvoiceResponse
	json.Unmarshal(w.Body.Bytes(), &list)
	found := false
	for _, item := range list.Data {
		if item.ID == invoice.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("invoice %d not found in filtered list (total=%d)", invoice.ID, list.Total)
	}

	// 12. Customer financial summary
	w = smokeJSON(t, r, http.MethodGet, "/customers/"+strconv.FormatUint(uint64(customerID), 10)+"/financial-summary", bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("financial-summary: %d %s", w.Code, w.Body.String())
	}
	var finWrap struct {
		Data dto.CustomerFinancialSummaryResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &finWrap)
	if finWrap.Data.TotalDebt != expectedTotal-partial {
		t.Fatalf("financial summary debt want %v got %v", expectedTotal-partial, finWrap.Data.TotalDebt)
	}
	if finWrap.Data.OpenInvoicesCount < 1 {
		t.Fatalf("expected open invoices count >= 1, got %d", finWrap.Data.OpenInvoicesCount)
	}

	// 13. Low-stock (stock 7, min 8)
	w = smokeJSON(t, r, http.MethodGet, "/inventory/low-stock", bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("low-stock: %d %s", w.Code, w.Body.String())
	}
	var low dto.PaginatedLowStockResponse
	json.Unmarshal(w.Body.Bytes(), &low)
	lowFound := false
	for _, p := range low.Products {
		if p.ID == product.ID {
			lowFound = true
			if p.StockQuantity != initialStock-qty {
				t.Fatalf("low-stock stock want %v got %v", initialStock-qty, p.StockQuantity)
			}
			if p.MissingQuantity != minStock-(initialStock-qty) {
				t.Fatalf("low-stock missing want %v got %v", minStock-(initialStock-qty), p.MissingQuantity)
			}
		}
	}
	if !lowFound {
		t.Fatal("product not found in low-stock list")
	}

	// 14. Report kao sef
	w = smokeJSON(t, r, http.MethodGet, "/reports/daily?date="+today, bossToken, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("daily report as sef: %d %s", w.Code, w.Body.String())
	}

	// 15. Radnik → report 403
	w = smokeJSON(t, r, http.MethodPost, "/users", bossToken, map[string]string{
		"username": "radnik1",
		"password": "password123",
		"role":     models.RoleWorker,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create worker: %d %s", w.Code, w.Body.String())
	}
	workerToken, _ := smokeLogin(t, r, "radnik1", "password123")
	w = smokeJSON(t, r, http.MethodGet, "/reports/daily?date="+today, workerToken, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("worker report want 403 got %d (%s)", w.Code, w.Body.String())
	}

	// 16. Storno računa
	debtBeforeCancel := dbCustomer.TotalDebt
	database.DB.First(&dbCustomer, customerID)
	debtBeforeCancel = dbCustomer.TotalDebt

	w = smokeJSON(t, r, http.MethodPut, "/invoices/"+strconv.FormatUint(uint64(invoice.ID), 10)+"/cancel", bossToken, map[string]string{
		"reason": "MVP smoke storno",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("cancel invoice: %d %s", w.Code, w.Body.String())
	}
	var cancelWrap struct {
		Data dto.CancelInvoiceResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &cancelWrap)
	cancelResp := cancelWrap.Data
	if cancelResp.CreatedByUser == nil || cancelResp.CreatedByUser.ID != bossID {
		t.Fatalf("cancel audit user want %d got %+v", bossID, cancelResp.CreatedByUser)
	}
	if cancelResp.Refund == nil {
		t.Fatal("expected refund for partially paid invoice")
	}
	if cancelResp.Refund.CreatedByUser == nil || cancelResp.Refund.CreatedByUser.ID != bossID {
		t.Fatalf("refund audit user want %d got %+v", bossID, cancelResp.Refund.CreatedByUser)
	}
	if cancelResp.Refund.Amount != partial {
		t.Fatalf("refund amount want %v got %v", partial, cancelResp.Refund.Amount)
	}

	database.DB.First(&dbInvoice, invoice.ID)
	if dbInvoice.Status != models.InvoiceStatusCancelled {
		t.Fatalf("invoice status after cancel want cancelled got %s", dbInvoice.Status)
	}
	database.DB.First(&dbProduct, product.ID)
	if dbProduct.StockQuantity != initialStock {
		t.Fatalf("stock after cancel want %v got %v", initialStock, dbProduct.StockQuantity)
	}
	database.DB.First(&dbCustomer, customerID)
	// remaining before cancel was expectedTotal-partial; cancel reduces debt by remaining → debt should be 0
	if dbCustomer.TotalDebt != 0 {
		t.Fatalf("customer debt after cancel want 0 (was %v before cancel), got %v", debtBeforeCancel, dbCustomer.TotalDebt)
	}
}

func mustBelgrade(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Belgrade")
	if err != nil {
		t.Fatalf("location: %v", err)
	}
	return loc
}
