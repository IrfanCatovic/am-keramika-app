package repositories

import (
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCategoryTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.ProductGroup{}, &models.Product{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func createTestCategory(t *testing.T, name, slug string, active bool) models.Category {
	t.Helper()
	category := models.Category{Name: name, Slug: slug}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
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

func TestCreateCategorySetsActiveByDefault(t *testing.T) {
	setupCategoryTestDB(t)

	category := &models.Category{Name: "Keramika", Slug: "keramika"}
	if err := CreateCategory(category); err != nil {
		t.Fatalf("create: %v", err)
	}
	if !category.IsActive {
		t.Fatal("expected new category to be active")
	}
}

func TestUpdateCategoryNameAndSlug(t *testing.T) {
	setupCategoryTestDB(t)

	category := createTestCategory(t, "Keramika", "keramika", true)
	category.Name = "Sanitarije"
	category.Slug = "sanitarije"
	if err := UpdateCategory(&category); err != nil {
		t.Fatalf("update: %v", err)
	}

	reloaded, err := GetCategoryByID("1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if reloaded.Name != "Sanitarije" || reloaded.Slug != "sanitarije" {
		t.Fatalf("unexpected category %+v", reloaded)
	}
}

func TestGetCategoriesDefaultExcludesInactive(t *testing.T) {
	setupCategoryTestDB(t)
	createTestCategory(t, "Aktivna", "aktivna", true)
	createTestCategory(t, "Neaktivna", "neaktivna", false)

	categories, err := GetCategories(false)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(categories) != 1 || categories[0].Slug != "aktivna" {
		t.Fatalf("expected only active category, got %+v", categories)
	}
}

func TestGetCategoriesIncludeInactive(t *testing.T) {
	setupCategoryTestDB(t)
	createTestCategory(t, "Aktivna", "aktivna", true)
	createTestCategory(t, "Neaktivna", "neaktivna", false)

	categories, err := GetCategories(true)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}
}

func TestDeleteEmptyCategory(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Prazna", "prazna", true)

	if err := DeleteCategory(category.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := GetCategoryByID("1")
	if err != ErrCategoryNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestDeleteCategoryWithGroupReturnsConflict(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Sa grupom", "sa-grupom", true)
	group := models.ProductGroup{Name: "Grupa", Slug: "grupa", CategoryID: category.ID}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}

	err := DeleteCategory(category.ID)
	if err != ErrCategoryHasGroupsOrProducts {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestDeleteCategoryWithProductReturnsConflict(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Sa proizvodom", "sa-proizvodom", true)
	product := models.Product{
		Name: "Pločica", Slug: "plocica", CategoryID: category.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	err := DeleteCategory(category.ID)
	if err != ErrCategoryHasGroupsOrProducts {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestDeactivateCategoryKeepsGroupsAndProducts(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Keramika", "keramika", true)
	group := models.ProductGroup{Name: "Grupa", Slug: "grupa", CategoryID: category.ID}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	product := models.Product{
		Name: "Pločica", Slug: "plocica", CategoryID: category.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	if err := UpdateCategoryStatus(category.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	var groupCount int64
	database.DB.Model(&models.ProductGroup{}).Where("category_id = ?", category.ID).Count(&groupCount)
	var productCount int64
	database.DB.Model(&models.Product{}).Where("category_id = ?", category.ID).Count(&productCount)
	if groupCount != 1 || productCount != 1 {
		t.Fatalf("expected group and product to remain, got groups=%d products=%d", groupCount, productCount)
	}
}

func TestReactivateCategoryReturnsToActiveList(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Keramika", "keramika", true)

	if err := UpdateCategoryStatus(category.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := UpdateCategoryStatus(category.ID, true); err != nil {
		t.Fatalf("reactivate: %v", err)
	}

	categories, err := GetCategories(false)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(categories) != 1 {
		t.Fatalf("expected reactivated category in active list, got %d", len(categories))
	}
}

func TestCreateProductRejectsInactiveCategory(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Neaktivna", "neaktivna", false)

	product := &models.Product{
		Name: "Pločica", Slug: "plocica", CategoryID: category.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
	}
	err := CreateProduct(product)
	if err != ErrCategoryInactive {
		t.Fatalf("expected inactive category error, got %v", err)
	}
}

func TestCreateProductGroupRejectsInactiveCategory(t *testing.T) {
	setupCategoryTestDB(t)
	category := createTestCategory(t, "Neaktivna", "neaktivna", false)

	group := &models.ProductGroup{Name: "Grupa", Slug: "grupa", CategoryID: category.ID}
	err := CreateProductGroup(group)
	if err != ErrCategoryInactive {
		t.Fatalf("expected inactive category error, got %v", err)
	}
}

func TestGetAllProductsExcludesInactiveCategoryProducts(t *testing.T) {
	setupCategoryTestDB(t)

	activeCat := createTestCategory(t, "Aktivna", "aktivna", true)
	inactiveCat := createTestCategory(t, "Neaktivna", "neaktivna", false)

	activeProduct := models.Product{
		Name: "Vidljiv", Slug: "vidljiv", CategoryID: activeCat.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
	}
	hiddenProduct := models.Product{
		Name: "Skriven", Slug: "skriven", CategoryID: inactiveCat.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
	}
	if err := database.DB.Create(&activeProduct).Error; err != nil {
		t.Fatalf("create active product: %v", err)
	}
	if err := database.DB.Create(&hiddenProduct).Error; err != nil {
		t.Fatalf("create hidden product: %v", err)
	}

	products, err := GetAllProducts("", "", false)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}
	if len(products) != 1 || products[0].Slug != "vidljiv" {
		t.Fatalf("expected only product from active category, got %+v", products)
	}

	allProducts, err := GetAllProducts("", "", true)
	if err != nil {
		t.Fatalf("list all products: %v", err)
	}
	if len(allProducts) != 2 {
		t.Fatalf("expected 2 products with includeInactive, got %d", len(allProducts))
	}
}
