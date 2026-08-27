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

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupProductPackageTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	database.DB = db

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.ProductGroup{},
		&models.Product{},
		&models.ProductImage{},
	); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
	os.Setenv("JWT_SECRET", "test-secret-package")
}

func setupProductPackageRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)

	public := r.Group("/public")
	{
		public.GET("/products", GetPublicProducts)
		public.GET("/products/:slug", GetPublicProductBySlug)
	}

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/products", CreateProduct)
			staff.GET("/products", GetAllProducts)
			staff.GET("/products/:id", GetProductById)
			staff.PUT("/products/:id", UpdateProduct)
		}
	}

	return r
}

func createPackageTestUser(t *testing.T, username, role string) models.User {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("user create: %v", err)
	}
	return user
}

func packageTestLogin(t *testing.T, r *gin.Engine, username string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Token == "" {
		t.Fatalf("bad token response: %s", w.Body.String())
	}
	return resp.Token
}

func createPackageTestCategory(t *testing.T, name, slug string) models.Category {
	t.Helper()
	cat := models.Category{
		Name:     name,
		Slug:     slug,
		IsActive: true,
	}
	if err := database.DB.Create(&cat).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	return cat
}

func TestProductPackageContract(t *testing.T) {
	setupProductPackageTestDB(t)
	r := setupProductPackageRouter()

	createPackageTestUser(t, "sef", models.RoleBoss)
	token := packageTestLogin(t, r, "sef")

	cat := createPackageTestCategory(t, "Pločice", "plocice")

	// 1. Normal product: saleByPackage=false -> existing behavior
	salePrice := 1500.0
	normalBody, _ := json.Marshal(dto.CreateProductRequest{
		Name:          "Standardne Pločice",
		CategoryID:    cat.ID,
		Unit:          "m²",
		SalePrice:     &salePrice,
		StockQuantity: 100,
		SaleByPackage: false,
	})
	req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(normalBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create normal product want 201 got %d: %s", w.Code, w.Body.String())
	}
	var normalResp dto.ProductResponse
	_ = json.Unmarshal(w.Body.Bytes(), &normalResp)
	if normalResp.SaleByPackage {
		t.Fatalf("expected saleByPackage=false")
	}
	if normalResp.PackageQuantity != 0 {
		t.Fatalf("expected packageQuantity=0, got %v", normalResp.PackageQuantity)
	}
	if normalResp.PackagePrice != nil {
		t.Fatalf("expected packagePrice=nil, got %v", *normalResp.PackagePrice)
	}

	// 2. Package product: saleByPackage=true, packageQuantity=2.70 -> valid
	pkgQty := 2.70
	packageBody, _ := json.Marshal(dto.CreateProductRequest{
		Name:            "Granitna Pločica Lux",
		CategoryID:      cat.ID,
		Unit:            "m²",
		SalePrice:       &salePrice, // 1500
		StockQuantity:   100,
		SaleByPackage:   true,
		PackageQuantity: &pkgQty,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(packageBody))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("create package product want 201 got %d: %s", w2.Code, w2.Body.String())
	}
	var pkgResp dto.ProductResponse
	_ = json.Unmarshal(w2.Body.Bytes(), &pkgResp)
	if !pkgResp.SaleByPackage {
		t.Fatalf("expected saleByPackage=true")
	}
	if pkgResp.PackageQuantity != 2.70 {
		t.Fatalf("expected packageQuantity=2.70, got %v", pkgResp.PackageQuantity)
	}
	// 1500 * 2.70 = 4050
	if pkgResp.PackagePrice == nil || *pkgResp.PackagePrice != 4050 {
		t.Fatalf("expected packagePrice=4050, got %v", pkgResp.PackagePrice)
	}

	// 3. Package enabled + quantity 0 -> reject (422)
	zeroQty := 0.0
	zeroPkgBody, _ := json.Marshal(dto.CreateProductRequest{
		Name:            "Neispravan Paket Nula",
		CategoryID:      cat.ID,
		Unit:            "m²",
		SalePrice:       &salePrice,
		StockQuantity:   100,
		SaleByPackage:   true,
		PackageQuantity: &zeroQty,
	})
	req3, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(zeroPkgBody))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnprocessableEntity {
		t.Fatalf("create package product with qty 0 want 422 got %d: %s", w3.Code, w3.Body.String())
	}

	// 4. Package enabled + negative -> reject (422)
	negQty := -2.70
	negPkgBody, _ := json.Marshal(dto.CreateProductRequest{
		Name:            "Neispravan Paket Negativno",
		CategoryID:      cat.ID,
		Unit:            "m²",
		SalePrice:       &salePrice,
		StockQuantity:   100,
		SaleByPackage:   true,
		PackageQuantity: &negQty,
	})
	req4, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(negPkgBody))
	req4.Header.Set("Authorization", "Bearer "+token)
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnprocessableEntity {
		t.Fatalf("create package product with negative qty want 422 got %d: %s", w4.Code, w4.Body.String())
	}

	// 5 & 6. Package price uses EFFECTIVE price:
	// regular 2350, discount 15% -> effective 2000, package 2.70 -> packagePrice 5400
	regPrice := 2350.0
	discount := 15.0
	salePackageBody, _ := json.Marshal(dto.CreateProductRequest{
		Name:            "Pločica Na Akciji",
		CategoryID:      cat.ID,
		Unit:            "m²",
		SalePrice:       &regPrice,
		StockQuantity:   100,
		IsOnSale:        true,
		DiscountPercent: &discount,
		SaleByPackage:   true,
		PackageQuantity: &pkgQty,
	})
	req5, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(salePackageBody))
	req5.Header.Set("Authorization", "Bearer "+token)
	req5.Header.Set("Content-Type", "application/json")
	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, req5)

	if w5.Code != http.StatusCreated {
		t.Fatalf("create sale package product want 201 got %d: %s", w5.Code, w5.Body.String())
	}
	var salePkgResp dto.ProductResponse
	_ = json.Unmarshal(w5.Body.Bytes(), &salePkgResp)
	if salePkgResp.EffectiveSalePrice != 2000 {
		t.Fatalf("effectiveSalePrice want 2000, got %v", salePkgResp.EffectiveSalePrice)
	}
	if salePkgResp.PackagePrice == nil || *salePkgResp.PackagePrice != 5400 {
		t.Fatalf("packagePrice want 5400, got %v", salePkgResp.PackagePrice)
	}

	// 10 & 11. Public catalog endpoint includes package fields, NO sensitive fields
	reqPub, _ := http.NewRequest(http.MethodGet, "/public/products/"+salePkgResp.Slug, nil)
	wPub := httptest.NewRecorder()
	r.ServeHTTP(wPub, reqPub)
	if wPub.Code != http.StatusOK {
		t.Fatalf("public product detail want 200 got %d: %s", wPub.Code, wPub.Body.String())
	}
	var pubResp dto.PublicProductResponse
	_ = json.Unmarshal(wPub.Body.Bytes(), &pubResp)
	if !pubResp.SaleByPackage {
		t.Fatalf("public expected saleByPackage=true")
	}
	if pubResp.PackageQuantity != 2.70 {
		t.Fatalf("public expected packageQuantity=2.70, got %v", pubResp.PackageQuantity)
	}
	if pubResp.PackagePrice == nil || *pubResp.PackagePrice != 5400 {
		t.Fatalf("public expected packagePrice=5400, got %v", pubResp.PackagePrice)
	}
	if pubResp.EffectiveSalePrice != 2000 {
		t.Fatalf("public expected effectiveSalePrice=2000, got %v", pubResp.EffectiveSalePrice)
	}

	// Also check raw JSON doesn't contain sensitive keys
	var rawPub map[string]interface{}
	_ = json.Unmarshal(wPub.Body.Bytes(), &rawPub)
	for _, sensitive := range []string{"purchasePrice", "marginPercent", "vatPercent", "stockQuantity"} {
		if _, exists := rawPub[sensitive]; exists {
			t.Fatalf("public DTO leaked sensitive key %q", sensitive)
		}
	}

	// Test Update: turning off saleByPackage resets packageQuantity to 0
	falseFlag := false
	updateBody, _ := json.Marshal(map[string]interface{}{
		"name":          salePkgResp.Name,
		"categoryID":    salePkgResp.CategoryID,
		"unit":          salePkgResp.Unit,
		"salePrice":     salePkgResp.SalePrice,
		"stockQuantity": salePkgResp.StockQuantity,
		"saleByPackage": falseFlag,
	})
	reqUpd, _ := http.NewRequest(http.MethodPut, "/products/"+strconv.FormatUint(uint64(salePkgResp.ID), 10), bytes.NewBuffer(updateBody))
	reqUpd.Header.Set("Authorization", "Bearer "+token)
	reqUpd.Header.Set("Content-Type", "application/json")
	wUpd := httptest.NewRecorder()
	r.ServeHTTP(wUpd, reqUpd)
	if wUpd.Code != http.StatusOK {
		t.Fatalf("update product want 200 got %d: %s", wUpd.Code, wUpd.Body.String())
	}
	var updResp dto.ProductResponse
	_ = json.Unmarshal(wUpd.Body.Bytes(), &updResp)
	if updResp.SaleByPackage {
		t.Fatalf("expected saleByPackage=false after update")
	}
	if updResp.PackageQuantity != 0 {
		t.Fatalf("expected packageQuantity=0 after turning off package sale, got %v", updResp.PackageQuantity)
	}
	if updResp.PackagePrice != nil {
		t.Fatalf("expected packagePrice=nil after turning off package sale")
	}
}
