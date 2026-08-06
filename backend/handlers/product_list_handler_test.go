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

func setupProductListHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.ProductGroup{}, &models.Product{}, &models.ProductImage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-product-list")
}

func setupProductListRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.GET("/products", GetAllProducts)
		}
	}
	return r
}

func productListToken(t *testing.T, r *gin.Engine, role string) string {
	t.Helper()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
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

func TestGetAllProductsGroupIDAndUngroupedConflict(t *testing.T) {
	setupProductListHandlerTestDB(t)
	r := setupProductListRouter()
	token := productListToken(t, r, models.RoleWorker)

	req := httptest.NewRequest(http.MethodGet, "/products?groupID=1&ungrouped=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestGetAllProductsInvalidPageLimit(t *testing.T) {
	setupProductListHandlerTestDB(t)
	r := setupProductListRouter()
	token := productListToken(t, r, models.RoleWorker)

	cases := []string{
		"/products?page=0",
		"/products?page=-1",
		"/products?limit=0",
		"/products?limit=101",
	}

	for _, path := range cases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d", path, w.Code)
		}
	}
}

func TestGetAllProductsPaginationResponseShape(t *testing.T) {
	setupProductListHandlerTestDB(t)
	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	for i := 0; i < 25; i++ {
		name := string(rune('A' + i))
		database.DB.Create(&models.Product{
			Name: name, Slug: "p-" + name, CategoryID: cat.ID,
			Unit: "kom", SalePrice: 10, IsActive: true,
		})
	}

	r := setupProductListRouter()
	token := productListToken(t, r, models.RoleManager)

	req := httptest.NewRequest(http.MethodGet, "/products?page=2&limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var resp dto.PaginatedProductListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Pagination.Page != 2 || resp.Pagination.Limit != 20 {
		t.Fatalf("unexpected pagination page/limit: %+v", resp.Pagination)
	}
	if resp.Pagination.TotalItems != 25 || resp.Pagination.TotalPages != 2 {
		t.Fatalf("unexpected totals: %+v", resp.Pagination)
	}
	if len(resp.Products) != 5 {
		t.Fatalf("expected 5 products on page 2, got %d", len(resp.Products))
	}
}

func TestGetAllProductsWorkerHidesSensitiveFields(t *testing.T) {
	setupProductListHandlerTestDB(t)
	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	purchase := 5.0
	margin := 20.0
	database.DB.Create(&models.Product{
		Name: "Pločica", Slug: "plocica", CategoryID: cat.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
		PurchasePrice: &purchase, MarginPercent: &margin,
	})

	r := setupProductListRouter()
	token := productListToken(t, r, models.RoleWorker)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if bytes.Contains(w.Body.Bytes(), []byte("purchasePrice")) || bytes.Contains(w.Body.Bytes(), []byte("marginPercent")) {
		t.Fatalf("worker must not see sensitive fields: %s", w.Body.String())
	}
}

func TestGetAllProductsPrimaryImageWithoutNPlusOne(t *testing.T) {
	setupProductListHandlerTestDB(t)
	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)

	var queryCount int
	database.DB.Callback().Query().Before("gorm:query").Register("count_queries_"+t.Name(), func(tx *gorm.DB) {
		queryCount++
	})

	products := make([]models.Product, 0, 5)
	for i := 0; i < 5; i++ {
		name := string(rune('A' + i))
		p := models.Product{Name: name, Slug: "p-" + name, CategoryID: cat.ID, Unit: "kom", SalePrice: 10, IsActive: true}
		database.DB.Create(&p)
		database.DB.Create(&models.ProductImage{ProductID: p.ID, URL: "u" + name, PublicID: "pid" + name, IsPrimary: true})
		products = append(products, p)
	}

	r := setupProductListRouter()
	token := productListToken(t, r, models.RoleWorker)

	queryCount = 0
	req := httptest.NewRequest(http.MethodGet, "/products?limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var resp dto.PaginatedProductListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Products) != 5 {
		t.Fatalf("expected 5 products, got %d", len(resp.Products))
	}
	for _, product := range resp.Products {
		if product.PrimaryImage == nil {
			t.Fatalf("expected primaryImage for product %d", product.ID)
		}
	}

	// List: count + products + batch primary images (+ auth/middleware queries already done).
	if queryCount > 8 {
		t.Fatalf("too many DB queries (%d), possible N+1 on primary images", queryCount)
	}
}
