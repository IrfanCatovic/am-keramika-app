package repositories

import (
	"strconv"
	"strings"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupHierarchyTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	err = db.AutoMigrate(&models.Category{}, &models.ProductGroup{}, &models.Product{}, &models.ProductImage{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	database.DB = db
}

func seedCategory(t *testing.T, name, slug string) models.Category {
	t.Helper()
	category := models.Category{Name: name, Slug: slug, IsActive: true}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	return category
}

func seedGroup(t *testing.T, name, slug string, categoryID uint) models.ProductGroup {
	t.Helper()
	group := models.ProductGroup{Name: name, Slug: slug, CategoryID: categoryID}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	return group
}

func seedProduct(t *testing.T, name, slug string, categoryID uint, groupID *uint) models.Product {
	t.Helper()
	product := models.Product{
		Name:       name,
		Slug:       slug,
		CategoryID: categoryID,
		GroupID:    groupID,
		Unit:       "kom",
		SalePrice:  10,
		IsActive:   true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return product
}

func TestUpdateProductGroup_RejectCategoryChangeWhenHasProducts(t *testing.T) {
	setupHierarchyTestDB(t)

	catA := seedCategory(t, "Keramika", "keramika")
	catB := seedCategory(t, "Grijanje", "grijanje")
	group := seedGroup(t, "Pločice", "plocice", catA.ID)
	groupID := group.ID
	seedProduct(t, "Pločica A", "plocica-a", catA.ID, &groupID)

	group.CategoryID = catB.ID
	err := UpdateProductGroup(&group)
	if err == nil {
		t.Fatal("expected error when changing category of group with products")
	}
	if !strings.Contains(err.Error(), "premjestite ili uklonite proizvode") {
		t.Fatalf("unexpected error: %v", err)
	}

	reloaded, err := GetProductGroupByID(group.ID)
	if err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if reloaded.CategoryID != catA.ID {
		t.Fatalf("category should remain %d, got %d", catA.ID, reloaded.CategoryID)
	}
}

func TestUpdateProductGroup_AllowCategoryChangeWhenEmpty(t *testing.T) {
	setupHierarchyTestDB(t)

	catA := seedCategory(t, "Keramika", "keramika")
	catB := seedCategory(t, "Grijanje", "grijanje")
	group := seedGroup(t, "Prazna", "prazna", catA.ID)

	group.Name = "Prazna 2"
	group.Slug = "prazna-2"
	group.CategoryID = catB.ID

	if err := UpdateProductGroup(&group); err != nil {
		t.Fatalf("expected empty group category change to succeed: %v", err)
	}

	reloaded, err := GetProductGroupByID(group.ID)
	if err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if reloaded.CategoryID != catB.ID {
		t.Fatalf("expected category %d, got %d", catB.ID, reloaded.CategoryID)
	}
}

func TestUpdateProduct_ClearGroupID(t *testing.T) {
	setupHierarchyTestDB(t)

	cat := seedCategory(t, "Keramika", "keramika")
	group := seedGroup(t, "Pločice", "plocice", cat.ID)
	groupID := group.ID
	product := seedProduct(t, "Proizvod", "proizvod", cat.ID, &groupID)

	product.GroupID = nil
	if err := UpdateProduct(&product); err != nil {
		t.Fatalf("clear group: %v", err)
	}

	reloaded, err := GetProductById(strconv.FormatUint(uint64(product.ID), 10))
	if err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if reloaded.GroupID != nil {
		t.Fatalf("expected group_id NULL, got %v", *reloaded.GroupID)
	}
}

func TestUpdateProduct_MoveToGroupSameCategory(t *testing.T) {
	setupHierarchyTestDB(t)

	cat := seedCategory(t, "Keramika", "keramika")
	groupA := seedGroup(t, "Grupa A", "grupa-a", cat.ID)
	groupB := seedGroup(t, "Grupa B", "grupa-b", cat.ID)
	groupAID := groupA.ID
	product := seedProduct(t, "Proizvod", "proizvod", cat.ID, &groupAID)

	groupBID := groupB.ID
	product.GroupID = &groupBID
	if err := UpdateProduct(&product); err != nil {
		t.Fatalf("move group: %v", err)
	}

	reloaded, err := GetProductById(strconv.FormatUint(uint64(product.ID), 10))
	if err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if reloaded.GroupID == nil || *reloaded.GroupID != groupB.ID {
		t.Fatalf("expected group %d, got %v", groupB.ID, reloaded.GroupID)
	}
	if reloaded.Group == nil || reloaded.Group.ID != groupB.ID {
		t.Fatalf("expected preloaded group %d", groupB.ID)
	}
}

func TestUpdateProduct_RejectGroupFromOtherCategory(t *testing.T) {
	setupHierarchyTestDB(t)

	catA := seedCategory(t, "Keramika", "keramika")
	catB := seedCategory(t, "Grijanje", "grijanje")
	groupB := seedGroup(t, "Radijatori", "radijatori", catB.ID)
	product := seedProduct(t, "Proizvod", "proizvod", catA.ID, nil)

	groupBID := groupB.ID
	product.GroupID = &groupBID
	err := UpdateProduct(&product)
	if err == nil {
		t.Fatal("expected rejection for group from other category")
	}
	if !strings.Contains(err.Error(), "grupa ne pripada izabranoj kategoriji") {
		t.Fatalf("unexpected error: %v", err)
	}
}
