package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInvoiceListHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Customer{}, &models.Invoice{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-invoice-list")
}

func setupInvoiceListRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.GET("/invoices", GetAllInvoices)
		}
	}
	return r
}

func invoiceListToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: "sef", PasswordHash: hash, Role: models.RoleBoss, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"username": "sef", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, _ := resp["token"].(string)
	return token
}

func TestGetAllInvoicesInvalidDateFormat(t *testing.T) {
	setupInvoiceListHandlerTestDB(t)
	r := setupInvoiceListRouter()
	token := invoiceListToken(t, r)

	req := httptest.NewRequest(http.MethodGet, "/invoices?fromDate=08-01-2026", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid fromDate, got %d (%s)", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/invoices?toDate=2026/08/01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid toDate, got %d", w.Code)
	}
}

func TestGetAllInvoicesToDateBeforeFromDate(t *testing.T) {
	setupInvoiceListHandlerTestDB(t)
	r := setupInvoiceListRouter()
	token := invoiceListToken(t, r)

	req := httptest.NewRequest(http.MethodGet, "/invoices?fromDate=2026-08-31&toDate=2026-08-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when toDate before fromDate, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAllInvoicesAcceptsPartiallyPaidStatus(t *testing.T) {
	setupInvoiceListHandlerTestDB(t)
	r := setupInvoiceListRouter()
	token := invoiceListToken(t, r)

	req := httptest.NewRequest(http.MethodGet, "/invoices?status=partially_paid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for partially_paid status, got %d (%s)", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/invoices?status=partiallyPaid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for legacy partiallyPaid, got %d", w.Code)
	}
}

func TestGetAllInvoicesResponseDTOShape(t *testing.T) {
	setupInvoiceListHandlerTestDB(t)
	r := setupInvoiceListRouter()
	token := invoiceListToken(t, r)

	customer := models.Customer{Name: "Dallas", Phone: "061", IsActive: true}
	database.DB.Create(&customer)
	invoice := models.Invoice{
		CreatedByUserID: 1,
		CustomerID:      &customer.ID,
		Status:          models.InvoiceStatusUnpaid,
		TotalAmount:     100,
		PaidAmount:      25,
	}
	database.DB.Create(&invoice)

	req := httptest.NewRequest(http.MethodGet, "/invoices?page=1&limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var resp dto.PaginatedInvoiceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Page != 1 || resp.Limit != 20 || resp.Total != 1 {
		t.Fatalf("unexpected pagination: %+v", resp)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 invoice")
	}
	item := resp.Data[0]
	if item.RemainingAmount != 75 {
		t.Fatalf("expected remainingAmount 75, got %v", item.RemainingAmount)
	}
	if item.CustomerID == nil || *item.CustomerID != customer.ID {
		t.Fatalf("expected customerID, got %v", item.CustomerID)
	}
	if item.Customer == nil || item.Customer.Name != "Dallas" {
		t.Fatalf("expected customer DTO, got %+v", item.Customer)
	}
}
