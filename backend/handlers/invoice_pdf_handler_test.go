package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInvoicePDFDB(t *testing.T) {
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
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-invoice-pdf")
	os.Setenv("COMPANY_NAME", "AM Keramika")
	os.Setenv("COMPANY_CITY", "Tutin")
}

func setupInvoicePDFRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.GET("/invoices/:id/pdf", handlers.GetInvoicePDF)
		}
	}
	return r
}

func seedPDFUser(t *testing.T) models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("pass12345"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	user := models.User{
		Username:     "pdfworker",
		PasswordHash: string(hash),
		Role:         models.RoleWorker,
		IsActive:     true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func loginPDF(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": "pdfworker", "password": "pass12345"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status %d body %s", w.Code, w.Body.String())
	}
	var resp dto.LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("login decode: %v", err)
	}
	return resp.Token
}

func seedPDFInvoice(t *testing.T, userID uint, status models.InvoiceStatus) models.Invoice {
	t.Helper()
	cat := models.Category{Name: "Keramika", Slug: "keramika-" + t.Name(), IsActive: true}
	if err := database.DB.Create(&cat).Error; err != nil {
		t.Fatalf("category: %v", err)
	}
	product := models.Product{
		Name: "Pločica Šćepan", CategoryID: cat.ID, Unit: "m2",
		SalePrice: 1500.5, StockQuantity: 100, IsActive: true, Slug: "p-" + t.Name(),
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}
	customer := models.Customer{Name: "Mujo Čengić", Phone: "061", IsActive: true}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("customer: %v", err)
	}
	cid := customer.ID
	invoice := models.Invoice{
		CreatedByUserID: userID,
		CustomerID:      &cid,
		TotalAmount:     3751.25,
		PaidAmount:      3751.25,
		Status:          status,
		Model:           gorm.Model{CreatedAt: time.Date(2026, 8, 6, 18, 0, 0, 0, time.UTC)},
	}
	if err := database.DB.Create(&invoice).Error; err != nil {
		t.Fatalf("invoice: %v", err)
	}
	item := models.InvoiceItem{
		InvoiceID: invoice.ID, ProductID: product.ID,
		Quantity: 2.5, UnitPrice: 1500.5, TotalPrice: 3751.25,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("item: %v", err)
	}
	return invoice
}

func TestGetInvoicePDF_OKAttachmentAndInline(t *testing.T) {
	setupInvoicePDFDB(t)
	user := seedPDFUser(t)
	invoice := seedPDFInvoice(t, user.ID, models.InvoiceStatusPaid)
	r := setupInvoicePDFRouter()
	token := loginPDF(t, r)

	req := httptest.NewRequest(http.MethodGet, "/invoices/"+strconv.FormatUint(uint64(invoice.ID), 10)+"/pdf?download=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/pdf") {
		t.Fatalf("content-type %q", ct)
	}
	if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF-")) {
		t.Fatal("missing PDF header")
	}
	cd := w.Header().Get("Content-Disposition")
	wantName := `filename="AM-Keramika-Racun-` + strconv.FormatUint(uint64(invoice.ID), 10) + `.pdf"`
	if !strings.Contains(cd, "attachment") || !strings.Contains(cd, wantName) {
		t.Fatalf("attachment disposition %q", cd)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/invoices/"+strconv.FormatUint(uint64(invoice.ID), 10)+"/pdf", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	cd2 := w2.Header().Get("Content-Disposition")
	if !strings.Contains(cd2, "inline") || !strings.Contains(cd2, wantName) {
		t.Fatalf("inline disposition %q", cd2)
	}
}

func TestGetInvoicePDF_NotFound(t *testing.T) {
	setupInvoicePDFDB(t)
	seedPDFUser(t)
	r := setupInvoicePDFRouter()
	token := loginPDF(t, r)

	req := httptest.NewRequest(http.MethodGet, "/invoices/99999/pdf", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d body %s", w.Code, w.Body.String())
	}
}

func TestGetInvoicePDF_Unauthorized(t *testing.T) {
	setupInvoicePDFDB(t)
	r := setupInvoicePDFRouter()
	req := httptest.NewRequest(http.MethodGet, "/invoices/1/pdf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("want 401/403 got %d", w.Code)
	}
}

func TestGetInvoicePDF_Cancelled(t *testing.T) {
	setupInvoicePDFDB(t)
	user := seedPDFUser(t)
	invoice := seedPDFInvoice(t, user.ID, models.InvoiceStatusCancelled)
	r := setupInvoicePDFRouter()
	token := loginPDF(t, r)

	req := httptest.NewRequest(http.MethodGet, "/invoices/"+strconv.FormatUint(uint64(invoice.ID), 10)+"/pdf?download=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF-")) {
		t.Fatal("missing PDF header")
	}
}
