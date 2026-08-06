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

func setupPricingHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Product{}, &models.ProductGroup{}, &models.ProductImage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-pricing")
}

func setupPricingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/products", CreateProduct)
			staff.PUT("/products/:id", UpdateProduct)
			staff.GET("/products/:id", GetProductById)
		}
	}
	return r
}

func pricingCreateUser(t *testing.T, username, role string) models.User {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: username, PasswordHash: hash, Role: role, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("user: %v", err)
	}
	return user
}

func pricingLogin(t *testing.T, r *gin.Engine, username string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login %d %s", w.Code, w.Body.String())
	}
	var resp dto.LoginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Token
}

func pricingSeedCategory(t *testing.T) models.Category {
	t.Helper()
	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	if err := database.DB.Create(&cat).Error; err != nil {
		t.Fatalf("category: %v", err)
	}
	return cat
}

func TestPricingCreateManualAndCalculated(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupPricingRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	token := pricingLogin(t, r, "sef")
	cat := pricingSeedCategory(t)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "Manual", "categoryID": cat.ID, "unit": "kom",
		"salePrice": 153, "stockQuantity": 1, "minStockQuantity": 0,
		"marginPercent": 0, "vatPercent": 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("manual create %d %s", w.Code, w.Body.String())
	}
	var manual dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &manual)
	if manual.SalePrice != 153 || manual.PricingMode != "manual" {
		t.Fatalf("manual resp %+v", manual)
	}

	body, _ = json.Marshal(map[string]interface{}{
		"name": "Auto", "categoryID": cat.ID, "unit": "kom",
		"purchasePrice": 100, "marginPercent": 25, "vatPercent": 20,
		"salePrice": 999, "stockQuantity": 1, "minStockQuantity": 0,
	})
	req = httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("calc create %d %s", w.Code, w.Body.String())
	}
	var calc dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &calc)
	if calc.SalePrice != 150 || calc.PricingMode != "calculated" {
		t.Fatalf("calc resp %+v", calc)
	}
}

func TestPricingUpdateRecalculatesAndModeSwitch(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupPricingRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	token := pricingLogin(t, r, "sef")
	cat := pricingSeedCategory(t)

	purchase := 100.0
	margin := 10.0
	vat := 10.0
	product := models.Product{
		Name: "P", Slug: "p", CategoryID: cat.ID, Unit: "kom",
		PurchasePrice: &purchase, MarginPercent: &margin, VatPercent: &vat,
		SalePrice: 130, StockQuantity: 1, IsActive: true,
	}
	if err := repositories.CreateProduct(&product); err != nil {
		t.Fatalf("seed: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"name": "P", "categoryID": cat.ID, "unit": "kom",
		"purchasePrice": 100, "marginPercent": 25, "vatPercent": 20,
		"stockQuantity": 1, "minStockQuantity": 0,
	})
	req := httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(int(product.ID)), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("recalc %d %s", w.Code, w.Body.String())
	}
	var updated dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.SalePrice != 150 {
		t.Fatalf("want 150 got %v", updated.SalePrice)
	}

	body, _ = json.Marshal(map[string]interface{}{
		"name": "P", "categoryID": cat.ID, "unit": "kom",
		"purchasePrice": 100, "marginPercent": 0, "vatPercent": 0,
		"stockQuantity": 1, "minStockQuantity": 0,
	})
	req = httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(int(product.ID)), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("calc->manual without salePrice want 422 got %d %s", w.Code, w.Body.String())
	}

	body, _ = json.Marshal(map[string]interface{}{
		"name": "P", "categoryID": cat.ID, "unit": "kom",
		"purchasePrice": 100, "marginPercent": 0, "vatPercent": 0,
		"salePrice": 153, "stockQuantity": 1, "minStockQuantity": 0,
	})
	req = httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(int(product.ID)), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("manual switch %d %s", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.SalePrice != 153 || updated.PricingMode != "manual" {
		t.Fatalf("manual switch %+v", updated)
	}
}

func TestWorkerPricingRules(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupPricingRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	pricingCreateUser(t, "radnik1", models.RoleWorker)
	bossToken := pricingLogin(t, r, "sef")
	workerToken := pricingLogin(t, r, "radnik1")
	cat := pricingSeedCategory(t)

	purchase := 100.0
	margin := 25.0
	vat := 20.0
	calcProduct := models.Product{
		Name: "Calc", Slug: "calc", CategoryID: cat.ID, Unit: "kom",
		PurchasePrice: &purchase, MarginPercent: &margin, VatPercent: &vat,
		SalePrice: 150, StockQuantity: 2, Description: "old", IsActive: true,
	}
	manualSale := 80.0
	manualProduct := models.Product{
		Name: "Manual", Slug: "manual", CategoryID: cat.ID, Unit: "kom",
		SalePrice: manualSale, StockQuantity: 2, IsActive: true,
	}
	zero := 0.0
	manualProduct.MarginPercent = &zero
	manualProduct.VatPercent = &zero
	repositories.CreateProduct(&calcProduct)
	repositories.CreateProduct(&manualProduct)

	// worker can change manual sale price
	body, _ := json.Marshal(map[string]interface{}{
		"name": "Manual", "categoryID": cat.ID, "unit": "kom",
		"salePrice": 90, "stockQuantity": 2, "minStockQuantity": 0,
	})
	req := httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(int(manualProduct.ID)), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+workerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("worker manual %d %s", w.Code, w.Body.String())
	}

	// worker cannot change calculated sale price
	body, _ = json.Marshal(map[string]interface{}{
		"name": "Calc", "categoryID": cat.ID, "unit": "kom",
		"salePrice": 200, "stockQuantity": 2, "minStockQuantity": 0, "description": "old",
	})
	req = httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(int(calcProduct.ID)), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+workerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("worker calc sale want 403 got %d %s", w.Code, w.Body.String())
	}

	// worker name update keeps hidden pricing
	body, _ = json.Marshal(map[string]interface{}{
		"name": "Calc Updated", "categoryID": cat.ID, "unit": "kom",
		"salePrice": 150, "stockQuantity": 2, "minStockQuantity": 0, "description": "new desc",
	})
	req = httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(int(calcProduct.ID)), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+workerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("worker rename %d %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/products/"+strconv.Itoa(int(calcProduct.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+bossToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var bossView dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &bossView)
	if bossView.Name != "Calc Updated" || bossView.Description != "new desc" {
		t.Fatalf("rename failed %+v", bossView)
	}
	if bossView.PurchasePrice == nil || *bossView.PurchasePrice != 100 {
		t.Fatalf("purchase wiped %+v", bossView.PurchasePrice)
	}
	if bossView.MarginPercent == nil || *bossView.MarginPercent != 25 {
		t.Fatalf("margin wiped %+v", bossView.MarginPercent)
	}
	if bossView.VatPercent == nil || *bossView.VatPercent != 20 {
		t.Fatalf("vat wiped %+v", bossView.VatPercent)
	}
	if bossView.SalePrice != 150 {
		t.Fatalf("sale changed %v", bossView.SalePrice)
	}

	req = httptest.NewRequest(http.MethodGet, "/products/"+strconv.Itoa(int(calcProduct.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+workerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var workerView dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &workerView)
	if workerView.PurchasePrice != nil || workerView.MarginPercent != nil || workerView.VatPercent != nil {
		t.Fatalf("worker leaked sensitive fields %+v", workerView)
	}
	if workerView.PricingMode != "calculated" {
		t.Fatalf("worker should see pricingMode")
	}
}

func TestCalculatedCreateWithoutPurchase(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupPricingRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	token := pricingLogin(t, r, "sef")
	cat := pricingSeedCategory(t)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "X", "categoryID": cat.ID, "unit": "kom",
		"marginPercent": 10, "stockQuantity": 0, "minStockQuantity": 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422 got %d %s", w.Code, w.Body.String())
	}
}
