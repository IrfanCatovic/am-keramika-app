package repositories

import (
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupLowStockTestDB(t *testing.T) {
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

func seedLowStockCategory(t *testing.T, name, slug string, active bool) models.Category {
	t.Helper()
	category := models.Category{Name: name, Slug: slug}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("category: %v", err)
	}
	if !active {
		if err := database.DB.Model(&category).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate category: %v", err)
		}
		category.IsActive = false
	} else {
		category.IsActive = true
	}
	return category
}

func seedLowStockProduct(t *testing.T, name, slug string, categoryID uint, groupID *uint, stock, min float64, active bool) models.Product {
	t.Helper()
	product := models.Product{
		Name:             name,
		Slug:             slug,
		CategoryID:       categoryID,
		GroupID:          groupID,
		Unit:             "kom",
		SalePrice:        10,
		StockQuantity:    stock,
		MinStockQuantity: min,
		IsActive:         true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}
	if !active {
		if err := database.DB.Model(&product).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate product: %v", err)
		}
		product.IsActive = false
	}
	return product
}

func TestListLowStockIncludesBelowMinimum(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Nisko", "nisko", cat.ID, nil, 2, 5, true)

	products, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(products) != 1 {
		t.Fatalf("expected below-min product, got total=%d len=%d", total, len(products))
	}
}

func TestListLowStockIncludesEqualToMinimum(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Jednak", "jednak", cat.ID, nil, 5, 5, true)

	products, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(products) != 1 {
		t.Fatalf("expected equal-min product, got total=%d", total)
	}
}

func TestListLowStockExcludesAboveMinimum(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Dovoljno", "dovoljno", cat.ID, nil, 6, 5, true)

	_, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected no products above min, got %d", total)
	}
}

func TestListLowStockIncludesZeroStockAndZeroMin(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Nula", "nula", cat.ID, nil, 0, 0, true)

	_, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 0/0 product included, got %d", total)
	}
}

func TestListLowStockExcludesInactiveProduct(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Neaktivan", "neaktivan", cat.ID, nil, 1, 5, false)

	_, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected inactive product excluded, got %d", total)
	}
}

func TestListLowStockExcludesInactiveCategory(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Neaktivna", "neaktivna", false)
	seedLowStockProduct(t, "Proizvod", "proizvod", cat.ID, nil, 1, 5, true)

	_, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected inactive category product excluded, got %d", total)
	}
}

func TestListLowStockSearch(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Verona Beige", "verona-beige", cat.ID, nil, 1, 5, true)
	seedLowStockProduct(t, "Drugi", "drugi", cat.ID, nil, 1, 5, true)

	products, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20, Search: "  VERONA "})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || products[0].Slug != "verona-beige" {
		t.Fatalf("expected search match, got total=%d %+v", total, products)
	}
}

func TestListLowStockCategoryAndGroupFilters(t *testing.T) {
	setupLowStockTestDB(t)
	catA := seedLowStockCategory(t, "Keramika", "keramika", true)
	catB := seedLowStockCategory(t, "Sanitarije", "sanitarije", true)
	group := models.ProductGroup{Name: "Verona", Slug: "verona", CategoryID: catA.ID}
	database.DB.Create(&group)

	seedLowStockProduct(t, "A", "a", catA.ID, &group.ID, 1, 5, true)
	seedLowStockProduct(t, "B", "b", catA.ID, nil, 1, 5, true)
	seedLowStockProduct(t, "C", "c", catB.ID, nil, 1, 5, true)

	byCat, totalCat, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20, CategoryID: "1"})
	if err != nil {
		t.Fatalf("category filter: %v", err)
	}
	if totalCat != 2 || len(byCat) != 2 {
		t.Fatalf("expected 2 in category, got total=%d", totalCat)
	}

	byGroup, totalGroup, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20, GroupID: "1"})
	if err != nil {
		t.Fatalf("group filter: %v", err)
	}
	if totalGroup != 1 || byGroup[0].Slug != "a" {
		t.Fatalf("expected group filter match, got total=%d %+v", totalGroup, byGroup)
	}
}

func TestListLowStockPaginationTotal(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	for i := 0; i < 5; i++ {
		seedLowStockProduct(t, "P"+string(rune('A'+i)), "p-"+string(rune('a'+i)), cat.ID, nil, 1, 5, true)
	}

	page1, total, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 5 || len(page1) != 2 {
		t.Fatalf("expected total 5 page len 2, got total=%d len=%d", total, len(page1))
	}
}

func TestListLowStockMostCriticalFirst(t *testing.T) {
	setupLowStockTestDB(t)
	cat := seedLowStockCategory(t, "Keramika", "keramika", true)
	seedLowStockProduct(t, "Blago", "blago", cat.ID, nil, 4, 5, true)   // -1
	seedLowStockProduct(t, "Kriticno", "kriticno", cat.ID, nil, 0, 5, true) // -5
	seedLowStockProduct(t, "Srednje", "srednje", cat.ID, nil, 2, 5, true)   // -3

	products, _, err := ListLowStockProducts(LowStockQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(products) != 3 {
		t.Fatalf("expected 3 products, got %d", len(products))
	}
	if products[0].Slug != "kriticno" || products[1].Slug != "srednje" || products[2].Slug != "blago" {
		t.Fatalf("unexpected sort order: %s, %s, %s", products[0].Slug, products[1].Slug, products[2].Slug)
	}
}
