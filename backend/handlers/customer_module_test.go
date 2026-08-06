package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

func setupCustomerModuleTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Customer{},
		&models.Category{},
		&models.Product{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Payment{},
		&models.PaymentAllocation{},
		&models.InventoryMovement{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-customer-module")
}

func setupCustomerModuleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/customers", CreateCustomer)
			staff.GET("/customers", GetAllCustomers)
			staff.PUT("/customers/:id", UpdateCustomer)
			staff.PUT("/customers/:id/status", UpdateCustomerStatus)
			staff.DELETE("/customers/:id", DeleteCustomer)
			staff.POST("/invoices", CreateInvoice)
		}
	}
	return r
}

func customerModuleToken(t *testing.T, r *gin.Engine) string {
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

func TestCreateInvoiceRejectsInactiveCustomer(t *testing.T) {
	setupCustomerModuleTestDB(t)
	r := setupCustomerModuleRouter()
	token := customerModuleToken(t, r)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Pločica", Slug: "plocica", CategoryID: cat.ID,
		Unit: "kom", SalePrice: 10, StockQuantity: 5, IsActive: true,
	}
	database.DB.Create(&product)

	customer := models.Customer{Name: "Neaktivan", Phone: "061111111"}
	database.DB.Create(&customer)
	database.DB.Model(&customer).Update("is_active", false)

	body, _ := json.Marshal(map[string]interface{}{
		"customerID": customer.ID,
		"items": []map[string]interface{}{
			{"productID": product.ID, "quantity": 1},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestUpdateCustomerHandler(t *testing.T) {
	setupCustomerModuleTestDB(t)
	r := setupCustomerModuleRouter()
	token := customerModuleToken(t, r)

	customer := models.Customer{Name: "Stari", Phone: "061111111", IsActive: true}
	database.DB.Create(&customer)

	body, _ := json.Marshal(dto.UpdateCustomerRequest{Name: "Novi", Phone: "062222222"})
	req := httptest.NewRequest(http.MethodPut, "/customers/"+strconv.FormatUint(uint64(customer.ID), 10), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAllCustomersPaginationShape(t *testing.T) {
	setupCustomerModuleTestDB(t)
	r := setupCustomerModuleRouter()
	token := customerModuleToken(t, r)

	for i := 0; i < 3; i++ {
		database.DB.Create(&models.Customer{Name: "Kupac " + strconv.Itoa(i), Phone: "06100000" + strconv.Itoa(i), IsActive: true})
	}

	req := httptest.NewRequest(http.MethodGet, "/customers?page=1&limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp dto.PaginatedCustomerResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Page != 1 || resp.Limit != 2 || resp.Total != 3 || resp.TotalPages != 2 {
		t.Fatalf("unexpected pagination: %+v", resp)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 items on page, got %d", len(resp.Data))
	}
}
