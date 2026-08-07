package handlers_test

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
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInventoryHandlerTestDB(t *testing.T) {
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
		&models.InventoryMovement{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-inventory-handler")
}

func setupInventoryHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleWorker))
		{
			staff.POST("/inventory/adjust", handlers.AdjustStock)
			staff.GET("/inventory/movements", handlers.GetInventoryMovements)
		}
	}
	return r
}

func inventoryHandlerToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: "invworker", PasswordHash: hash, Role: models.RoleWorker, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"username": "invworker", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, _ := resp["token"].(string)
	return token
}

func seedInventoryHandlerProduct(t *testing.T, stock float64) *models.Product {
	t.Helper()
	category := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	product := models.Product{
		Name:             "Pločica",
		Slug:             "plocica",
		CategoryID:       category.ID,
		Unit:             "m2",
		SalePrice:        1000,
		StockQuantity:    stock,
		MinStockQuantity: 5,
		IsActive:         true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return &product
}

func TestAdjustStockHandlerUsesNewQuantity(t *testing.T) {
	setupInventoryHandlerTestDB(t)
	r := setupInventoryHandlerRouter()
	token := inventoryHandlerToken(t, r)
	product := seedInventoryHandlerProduct(t, 20)

	payload, _ := json.Marshal(dto.AdjustStockRequest{
		ProductID:   product.ID,
		NewQuantity: 25,
		Note:        "Popis",
	})
	req := httptest.NewRequest(http.MethodPost, "/inventory/adjust", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}

	var updated models.Product
	database.DB.First(&updated, product.ID)
	if updated.StockQuantity != 25 {
		t.Fatalf("stock = %v want 25", updated.StockQuantity)
	}
}

func TestAdjustStockHandlerRejectsNegativeQuantity(t *testing.T) {
	setupInventoryHandlerTestDB(t)
	r := setupInventoryHandlerRouter()
	token := inventoryHandlerToken(t, r)
	product := seedInventoryHandlerProduct(t, 10)

	payload, _ := json.Marshal(map[string]interface{}{
		"productID":   product.ID,
		"newQuantity": -1,
	})
	req := httptest.NewRequest(http.MethodPost, "/inventory/adjust", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d want 400", w.Code)
	}
}

func TestGetInventoryMovementsPagination(t *testing.T) {
	setupInventoryHandlerTestDB(t)
	r := setupInventoryHandlerRouter()
	token := inventoryHandlerToken(t, r)
	product := seedInventoryHandlerProduct(t, 10)

	user := models.User{}
	if err := database.DB.Where("username = ?", "invworker").First(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if _, err := repositories.AdjustStock(product.ID, 12, "A", user.ID); err != nil {
		t.Fatalf("adjust: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/inventory/movements?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}

	var resp dto.PaginatedInventoryMovementsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Movements) != 1 {
		t.Fatalf("movements = %d want 1", len(resp.Movements))
	}
}
