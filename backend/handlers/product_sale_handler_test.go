package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/middleware"

	"github.com/gin-gonic/gin"
)

func setupSaleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)
	public := r.Group("/public")
	{
		public.GET("/products", GetPublicProducts)
		public.GET("/products/:slug", GetPublicProductBySlug)
		public.GET("/categories", GetPublicCategories)
		public.GET("/product-groups", GetPublicProductGroups)
	}
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

func TestProductSaleCreateAndEffectivePrice(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	token := pricingLogin(t, r, "sef")
	cat := pricingSeedCategory(t)

	discount := 15.0
	body, _ := json.Marshal(map[string]interface{}{
		"name": "Akcija", "categoryID": cat.ID, "unit": "kom",
		"salePrice": 2350, "stockQuantity": 5, "minStockQuantity": 0,
		"marginPercent": 0, "vatPercent": 0,
		"isOnSale": true, "discountPercent": discount,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create %d %s", w.Code, w.Body.String())
	}
	var created dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.SalePrice != 2350 {
		t.Fatalf("salePrice want 2350 got %v", created.SalePrice)
	}
	if created.EffectiveSalePrice != 2000 {
		t.Fatalf("effective want 2000 got %v", created.EffectiveSalePrice)
	}
	if created.DiscountPercent != 15 || !created.IsOnSale {
		t.Fatalf("sale flags %+v", created)
	}
}

func TestProductSaleRejectsInvalidDiscount(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	token := pricingLogin(t, r, "sef")
	cat := pricingSeedCategory(t)

	cases := []struct {
		name     string
		onSale   bool
		discount float64
	}{
		{"neg", false, -1},
		{"ge100", false, 100},
		{"saleZero", true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]interface{}{
				"name": "Bad-" + tc.name, "categoryID": cat.ID, "unit": "kom",
				"salePrice": 1000, "stockQuantity": 1, "minStockQuantity": 0,
				"marginPercent": 0, "vatPercent": 0,
				"isOnSale": tc.onSale, "discountPercent": tc.discount,
			})
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("want 422 got %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestWorkerCannotSetDiscountPercent(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	pricingCreateUser(t, "radnik", models.RoleWorker)
	token := pricingLogin(t, r, "radnik")
	cat := pricingSeedCategory(t)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "WorkerSale", "categoryID": cat.ID, "unit": "kom",
		"salePrice": 1000, "stockQuantity": 1, "minStockQuantity": 0,
		"isOnSale": true, "discountPercent": 10,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %s", w.Code, w.Body.String())
	}
}

func TestPublicProductOmitsSensitiveFields(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	pricingCreateUser(t, "sef", models.RoleBoss)
	token := pricingLogin(t, r, "sef")
	cat := pricingSeedCategory(t)

	purchase := 500.0
	margin := 20.0
	vat := 20.0
	body, _ := json.Marshal(map[string]interface{}{
		"name": "PublicSafe", "categoryID": cat.ID, "unit": "kom",
		"purchasePrice": purchase, "marginPercent": margin, "vatPercent": vat,
		"stockQuantity": 3, "minStockQuantity": 0,
		"isOnSale": true, "discountPercent": 10,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create %d %s", w.Code, w.Body.String())
	}
	var created dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	req = httptest.NewRequest(http.MethodGet, "/public/products/"+created.Slug, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("public get %d %s", w.Code, w.Body.String())
	}
	raw := w.Body.String()
	if strings.Contains(raw, "purchasePrice") || strings.Contains(raw, "marginPercent") || strings.Contains(raw, "vatPercent") {
		t.Fatalf("public body leaked sensitive fields: %s", raw)
	}
	if strings.Contains(raw, "stockQuantity") {
		t.Fatalf("public body should not expose stockQuantity: %s", raw)
	}
	var pub dto.PublicProductResponse
	json.Unmarshal(w.Body.Bytes(), &pub)
	if !pub.InStock {
		t.Fatalf("expected inStock true")
	}
	if pub.EffectiveSalePrice <= 0 || pub.SalePrice <= 0 {
		t.Fatalf("prices missing %+v", pub)
	}

	req = httptest.NewRequest(http.MethodGet, "/products/"+strconv.FormatUint(uint64(created.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var staff dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &staff)
	if staff.StockQuantity != 3 {
		t.Fatalf("staff stock want 3 got %v", staff.StockQuantity)
	}
}
