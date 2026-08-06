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

func setupMinStockHandlerTestDB(t *testing.T) {
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
	os.Setenv("JWT_SECRET", "test-secret-min-stock")
}

func setupMinStockRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/products", CreateProduct)
			staff.PUT("/products/:id", UpdateProduct)
			staff.GET("/products/:id", GetProductById)
		}
	}
	return r
}

func minStockToken(t *testing.T, r *gin.Engine) (string, uint) {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: "sef", PasswordHash: hash, Role: models.RoleBoss, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("user: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"username": "sef", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp dto.LoginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Token, user.ID
}

func TestCreateProductWithMinStockQuantity(t *testing.T) {
	setupMinStockHandlerTestDB(t)
	r := setupMinStockRouter()
	token, _ := minStockToken(t, r)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)

	body, _ := json.Marshal(map[string]interface{}{
		"name":             "Verona",
		"categoryID":       cat.ID,
		"unit":             "m2",
		"salePrice":        10,
		"stockQuantity":    3,
		"minStockQuantity": 5,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	var resp dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.MinStockQuantity != 5 {
		t.Fatalf("expected minStockQuantity 5, got %v", resp.MinStockQuantity)
	}
}

func TestCreateProductMinStockDefaultsToZero(t *testing.T) {
	setupMinStockHandlerTestDB(t)
	r := setupMinStockRouter()
	token, _ := minStockToken(t, r)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)

	body, _ := json.Marshal(map[string]interface{}{
		"name":          "BezMin",
		"categoryID":    cat.ID,
		"unit":          "kom",
		"salePrice":     10,
		"stockQuantity": 1,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	var resp dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.MinStockQuantity != 0 {
		t.Fatalf("expected default minStockQuantity 0, got %v", resp.MinStockQuantity)
	}
}

func TestCreateProductRejectsNegativeMinStock(t *testing.T) {
	setupMinStockHandlerTestDB(t)
	r := setupMinStockRouter()
	token, _ := minStockToken(t, r)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)

	body, _ := json.Marshal(map[string]interface{}{
		"name":             "Neg",
		"categoryID":       cat.ID,
		"unit":             "kom",
		"salePrice":        10,
		"stockQuantity":    1,
		"minStockQuantity": -1,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative minStockQuantity, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestUpdateProductMinStockQuantity(t *testing.T) {
	setupMinStockHandlerTestDB(t)
	r := setupMinStockRouter()
	token, _ := minStockToken(t, r)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "P", Slug: "p", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 10, StockQuantity: 5, MinStockQuantity: 1, IsActive: true,
	}
	database.DB.Create(&product)

	body, _ := json.Marshal(map[string]interface{}{
		"name":             "P",
		"categoryID":       cat.ID,
		"unit":             "kom",
		"salePrice":        10,
		"stockQuantity":    5,
		"minStockQuantity": 7,
	})
	req := httptest.NewRequest(http.MethodPut, "/products/"+strconv.FormatUint(uint64(product.ID), 10), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var resp dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.MinStockQuantity != 7 {
		t.Fatalf("expected updated minStockQuantity 7, got %v", resp.MinStockQuantity)
	}
}
