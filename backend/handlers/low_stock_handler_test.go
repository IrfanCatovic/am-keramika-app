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

func setupLowStockHandlerTestDB(t *testing.T) {
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
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-low-stock")
}

func setupLowStockRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.GET("/inventory/low-stock", GetLowStock)
		}
	}
	return r
}

func lowStockToken(t *testing.T, r *gin.Engine, role string) string {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: "user1", PasswordHash: hash, Role: role, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"username": "user1", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, _ := resp["token"].(string)
	return token
}

func TestGetLowStockMissingQuantity(t *testing.T) {
	setupLowStockHandlerTestDB(t)
	r := setupLowStockRouter()
	token := lowStockToken(t, r, models.RoleManager)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Verona", Slug: "verona", CategoryID: cat.ID, Unit: "m2",
		SalePrice: 10, StockQuantity: 2, MinStockQuantity: 5, IsActive: true,
	}
	database.DB.Create(&product)

	req := httptest.NewRequest(http.MethodGet, "/inventory/low-stock", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var resp dto.PaginatedLowStockResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(resp.Products))
	}
	if resp.Products[0].MissingQuantity != 3 {
		t.Fatalf("expected missingQuantity 3, got %v", resp.Products[0].MissingQuantity)
	}
	if resp.Products[0].Group != nil {
		t.Fatal("expected group null")
	}
}

func TestGetLowStockWorkerHasNoSensitiveFields(t *testing.T) {
	setupLowStockHandlerTestDB(t)
	r := setupLowStockRouter()
	token := lowStockToken(t, r, models.RoleWorker)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	purchase := 5.0
	margin := 20.0
	product := models.Product{
		Name: "Verona", Slug: "verona", CategoryID: cat.ID, Unit: "m2",
		SalePrice: 10, StockQuantity: 1, MinStockQuantity: 5, IsActive: true,
		PurchasePrice: &purchase, MarginPercent: &margin,
	}
	database.DB.Create(&product)

	req := httptest.NewRequest(http.MethodGet, "/inventory/low-stock", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if bytes.Contains(w.Body.Bytes(), []byte("purchasePrice")) || bytes.Contains(w.Body.Bytes(), []byte("marginPercent")) {
		t.Fatalf("worker must not receive sensitive fields: %s", w.Body.String())
	}
}

func TestGetLowStockPrimaryImageWithoutNPlusOne(t *testing.T) {
	setupLowStockHandlerTestDB(t)

	var queryCount int
	database.DB.Callback().Query().Before("gorm:query").Register("count_queries_"+t.Name(), func(tx *gorm.DB) {
		queryCount++
	})

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	for i := 0; i < 5; i++ {
		name := string(rune('A' + i))
		p := models.Product{
			Name: name, Slug: "p-" + name, CategoryID: cat.ID, Unit: "kom",
			SalePrice: 10, StockQuantity: 1, MinStockQuantity: 5, IsActive: true,
		}
		database.DB.Create(&p)
		database.DB.Create(&models.ProductImage{ProductID: p.ID, URL: "u" + name, PublicID: "pid" + name, IsPrimary: true})
	}

	r := setupLowStockRouter()
	token := lowStockToken(t, r, models.RoleWorker)

	queryCount = 0
	req := httptest.NewRequest(http.MethodGet, "/inventory/low-stock?limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var resp dto.PaginatedLowStockResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Products) != 5 {
		t.Fatalf("expected 5 products, got %d", len(resp.Products))
	}
	for _, product := range resp.Products {
		if product.PrimaryImage == nil {
			t.Fatalf("expected primaryImage for product %d", product.ID)
		}
	}
	if queryCount > 8 {
		t.Fatalf("too many DB queries (%d), possible N+1 on primary images", queryCount)
	}
}
