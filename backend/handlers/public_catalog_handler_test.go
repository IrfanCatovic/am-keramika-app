package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
)

func seedPublicCatalog(t *testing.T) (models.Category, models.Category, models.ProductGroup, models.Product, models.Product) {
	t.Helper()
	activeCat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	inactiveCat := models.Category{Name: "Stara", Slug: "stara", IsActive: true}
	database.DB.Create(&activeCat)
	database.DB.Create(&inactiveCat)
	database.DB.Model(&inactiveCat).Update("is_active", false)

	group := models.ProductGroup{Name: "Antila", Slug: "antila", CategoryID: activeCat.ID}
	database.DB.Create(&group)

	featured := models.Product{
		Name: "Pločica A", Slug: "plocica-a", CategoryID: activeCat.ID, GroupID: &group.ID,
		Unit: "m2", SalePrice: 2350, StockQuantity: 10, IsActive: true,
		IsOnSale: true, DiscountPercent: 15, ShowOnHomepage: true,
	}
	regular := models.Product{
		Name: "Pločica B", Slug: "plocica-b", CategoryID: activeCat.ID, GroupID: &group.ID,
		Unit: "m2", SalePrice: 1800, StockQuantity: 0, IsActive: true,
	}
	inactive := models.Product{
		Name: "Sakriven", Slug: "sakriven", CategoryID: activeCat.ID,
		Unit: "kom", SalePrice: 100, StockQuantity: 5, IsActive: true,
	}
	orphanInactiveCat := models.Product{
		Name: "U staroj", Slug: "u-staroj", CategoryID: inactiveCat.ID,
		Unit: "kom", SalePrice: 200, StockQuantity: 5, IsActive: true,
	}
	database.DB.Create(&featured)
	database.DB.Create(&regular)
	database.DB.Create(&inactive)
	database.DB.Create(&orphanInactiveCat)
	database.DB.Model(&inactive).Update("is_active", false)
	return activeCat, inactiveCat, group, featured, regular
}

func TestPublicCatalogFiltersAndSafety(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	r.GET("/public/categories/:slug", GetPublicCategoryBySlug)
	activeCat, _, group, featured, regular := seedPublicCatalog(t)

	t.Run("only active products", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Pagination.TotalItems != 2 {
			t.Fatalf("want 2 active products got %d", resp.Pagination.TotalItems)
		}
		for _, p := range resp.Products {
			if p.Slug == "sakriven" || p.Slug == "u-staroj" {
				t.Fatalf("inactive leaked: %s", p.Slug)
			}
			raw, _ := json.Marshal(p)
			s := string(raw)
			if strings.Contains(s, "purchasePrice") || strings.Contains(s, "marginPercent") || strings.Contains(s, "stockQuantity") {
				t.Fatalf("sensitive field in public DTO: %s", s)
			}
		}
	})

	t.Run("category slug filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?categorySlug=keramika", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Pagination.TotalItems != 2 {
			t.Fatalf("got %d", resp.Pagination.TotalItems)
		}
		_ = activeCat
	})

	t.Run("group slug filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?categorySlug=keramika&groupSlug=antila", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Pagination.TotalItems != 2 {
			t.Fatalf("got %d", resp.Pagination.TotalItems)
		}
		_ = group
	})

	t.Run("search", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?search=plocica-a", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Pagination.TotalItems != 1 || resp.Products[0].Slug != "plocica-a" {
			t.Fatalf("search failed %+v", resp)
		}
	})

	t.Run("onSale", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?onSale=true", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Products) != 1 || !resp.Products[0].IsOnSale {
			t.Fatalf("onSale %+v", resp.Products)
		}
		if resp.Products[0].EffectiveSalePrice != 2000 {
			t.Fatalf("effective want 2000 got %v", resp.Products[0].EffectiveSalePrice)
		}
	})

	t.Run("homepage featured", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?homepage=true", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Products) != 1 || !resp.Products[0].ShowOnHomepage {
			t.Fatalf("featured %+v", resp.Products)
		}
	})

	t.Run("inStock", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?inStock=true", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Products) != 1 || !resp.Products[0].InStock {
			t.Fatalf("inStock %+v", resp.Products)
		}
	})

	t.Run("excludeId related", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?groupSlug=antila&excludeId="+strconv.FormatUint(uint64(featured.ID), 10), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Products) != 1 || resp.Products[0].ID != regular.ID {
			t.Fatalf("exclude related %+v", resp.Products)
		}
	})

	t.Run("sort price asc", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?sort=price_asc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.PaginatedPublicProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Products) < 2 {
			t.Fatalf("need 2 products")
		}
		if resp.Products[0].EffectiveSalePrice > resp.Products[1].EffectiveSalePrice {
			t.Fatalf("price_asc order wrong: %v then %v", resp.Products[0].EffectiveSalePrice, resp.Products[1].EffectiveSalePrice)
		}
	})

	t.Run("category by slug", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/categories/keramika", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
	})

	t.Run("random returns without error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public/products?random=true&limit=8", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
	})
}
