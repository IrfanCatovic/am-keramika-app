package repositories

import (
	"fmt"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupProductListTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.ProductGroup{}, &models.Product{}, &models.ProductImage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedProductListItem(t *testing.T, name, slug string, categoryID uint, groupID *uint, active bool) models.Product {
	t.Helper()
	product := models.Product{
		Name: name, Slug: slug, CategoryID: categoryID, GroupID: groupID,
		Unit: "kom", SalePrice: 10, IsActive: active,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product %s: %v", name, err)
	}
	if !active {
		if err := database.DB.Model(&product).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate product: %v", err)
		}
		product.IsActive = false
	}
	return product
}

func TestListProductsGroupIDFilter(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	groupA := models.ProductGroup{Name: "A", Slug: "a", CategoryID: cat.ID}
	groupB := models.ProductGroup{Name: "B", Slug: "b", CategoryID: cat.ID}
	database.DB.Create(&groupA)
	database.DB.Create(&groupB)

	seedProductListItem(t, "Alpha", "alpha", cat.ID, &groupA.ID, true)
	seedProductListItem(t, "Beta", "beta", cat.ID, &groupB.ID, true)

	products, total, err := ListProducts(ProductListQuery{
		GroupID: "1",
		Page:    1,
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(products) != 1 || products[0].Slug != "alpha" {
		t.Fatalf("expected one product in group A, got total=%d %+v", total, products)
	}
}

func TestListProductsUngroupedFilter(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	group := models.ProductGroup{Name: "Grupa", Slug: "grupa", CategoryID: cat.ID}
	database.DB.Create(&group)

	seedProductListItem(t, "Bez grupe", "bez-grupe", cat.ID, nil, true)
	seedProductListItem(t, "U grupi", "u-grupi", cat.ID, &group.ID, true)

	products, total, err := ListProducts(ProductListQuery{
		Ungrouped: true,
		Page:      1,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(products) != 1 || products[0].Slug != "bez-grupe" {
		t.Fatalf("expected ungrouped product, got total=%d %+v", total, products)
	}
}

func TestListProductsCombinedFilters(t *testing.T) {
	setupProductListTestDB(t)
	catA := createTestCategory(t, "Keramika", "keramika", true)
	catB := createTestCategory(t, "Sanitarije", "sanitarije", true)
	group := models.ProductGroup{Name: "Grupa", Slug: "grupa", CategoryID: catA.ID}
	database.DB.Create(&group)

	seedProductListItem(t, "Verona A", "verona-a", catA.ID, &group.ID, true)
	seedProductListItem(t, "Verona B", "verona-b", catB.ID, nil, true)
	seedProductListItem(t, "Drugo", "drugo", catA.ID, nil, true)

	products, total, err := ListProducts(ProductListQuery{
		Search:     "verona",
		CategoryID: "1",
		GroupID:    "1",
		Page:       1,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(products) != 1 || products[0].Slug != "verona-a" {
		t.Fatalf("expected combined filter match, got total=%d %+v", total, products)
	}
}

func TestListProductsPaginationTotals(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	for i := 1; i <= 25; i++ {
		name := fmt.Sprintf("Product %02d", i)
		seedProductListItem(t, name, fmt.Sprintf("product-%02d", i), cat.ID, nil, true)
	}

	_, total, err := ListProducts(ProductListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 25 {
		t.Fatalf("expected totalItems 25, got %d", total)
	}

	page2, total2, err := ListProducts(ProductListQuery{Page: 2, Limit: 20})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if total2 != 25 || len(page2) != 5 {
		t.Fatalf("expected 5 items on page 2, got total=%d len=%d", total2, len(page2))
	}
}

func TestListProductsStableSort(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	seedProductListItem(t, "Beta", "beta", cat.ID, nil, true)
	seedProductListItem(t, "Alpha", "alpha", cat.ID, nil, true)
	seedProductListItem(t, "Alpha", "alpha-2", cat.ID, nil, true)

	products, _, err := ListProducts(ProductListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(products) != 3 {
		t.Fatalf("expected 3 products, got %d", len(products))
	}
	if products[0].Slug != "alpha" || products[1].Slug != "alpha-2" || products[2].Slug != "beta" {
		t.Fatalf("unexpected sort order: %+v", []string{products[0].Slug, products[1].Slug, products[2].Slug})
	}
}

func TestListProductsIncludeInactive(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	seedProductListItem(t, "Aktivan", "aktivan", cat.ID, nil, true)
	seedProductListItem(t, "Neaktivan", "neaktivan", cat.ID, nil, false)

	activeOnly, totalActive, err := ListProducts(ProductListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("active list: %v", err)
	}
	if totalActive != 1 || len(activeOnly) != 1 {
		t.Fatalf("expected 1 active product, got total=%d len=%d", totalActive, len(activeOnly))
	}

	all, totalAll, err := ListProducts(ProductListQuery{IncludeInactive: true, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("all list: %v", err)
	}
	if totalAll != 2 || len(all) != 2 {
		t.Fatalf("expected 2 products with includeInactive, got total=%d len=%d", totalAll, len(all))
	}
}

func TestListProductsSearchMatchesSlug(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	seedProductListItem(t, "Pločica", "verona-classic", cat.ID, nil, true)
	seedProductListItem(t, "Druga", "druga", cat.ID, nil, true)

	products, total, err := ListProducts(ProductListQuery{
		Search: "  VERONA ",
		Page:   1,
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(products) != 1 || products[0].Slug != "verona-classic" {
		t.Fatalf("expected slug search match, got total=%d %+v", total, products)
	}
}

func TestGetPrimaryImagesForProductsBatch(t *testing.T) {
	setupProductListTestDB(t)
	cat := createTestCategory(t, "Keramika", "keramika", true)
	p1 := seedProductListItem(t, "A", "a", cat.ID, nil, true)
	p2 := seedProductListItem(t, "B", "b", cat.ID, nil, true)

	database.DB.Create(&models.ProductImage{ProductID: p1.ID, URL: "u1", PublicID: "p1", IsPrimary: true})
	database.DB.Create(&models.ProductImage{ProductID: p2.ID, URL: "u2", PublicID: "p2", IsPrimary: true})

	images, err := GetPrimaryImagesForProducts([]uint{p1.ID, p2.ID})
	if err != nil {
		t.Fatalf("batch primary: %v", err)
	}
	if len(images) != 2 || images[p1.ID].URL != "u1" || images[p2.ID].URL != "u2" {
		t.Fatalf("unexpected batch result: %+v", images)
	}
}
