package repositories_test

import (
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInventoryTestDB(t *testing.T) {
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
		&models.InventoryMovement{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedInventoryProduct(t *testing.T, stock float64) (*models.Product, *models.User) {
	t.Helper()
	user := models.User{Username: "worker1", PasswordHash: "hash", Role: models.RoleWorker, IsActive: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	category := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	product := models.Product{
		Name:             "Test pločica",
		Slug:             "test-plocica",
		CategoryID:       category.ID,
		Unit:             "m2",
		SalePrice:        1000,
		StockQuantity:    stock,
		MinStockQuantity: 5,
		IsActive:         true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return &product, &user
}

func TestAdjustStockIncrease(t *testing.T) {
	setupInventoryTestDB(t)
	product, user := seedInventoryProduct(t, 20)

	result, err := repositories.AdjustStock(product.ID, 25, "Fizički popis", user.ID)
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if result.PreviousStock != 20 || result.NewStock != 25 || result.Change != 5 {
		t.Fatalf("unexpected result: %+v", result)
	}

	var updated models.Product
	if err := database.DB.First(&updated, product.ID).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if updated.StockQuantity != 25 {
		t.Fatalf("stock = %v want 25", updated.StockQuantity)
	}

	var movement models.InventoryMovement
	if err := database.DB.Where("product_id = ?", product.ID).First(&movement).Error; err != nil {
		t.Fatalf("movement: %v", err)
	}
	if movement.MovementType != "adjust" || movement.Quantity != 5 || movement.Note != "Fizički popis" {
		t.Fatalf("unexpected movement: %+v", movement)
	}
	if movement.CreatedByUserID != user.ID {
		t.Fatalf("createdBy = %v want %v", movement.CreatedByUserID, user.ID)
	}
}

func TestAdjustStockDecrease(t *testing.T) {
	setupInventoryTestDB(t)
	product, user := seedInventoryProduct(t, 25)

	result, err := repositories.AdjustStock(product.ID, 18, "Oštećena roba", user.ID)
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if result.Change != -7 {
		t.Fatalf("change = %v want -7", result.Change)
	}
}

func TestAdjustStockToZero(t *testing.T) {
	setupInventoryTestDB(t)
	product, user := seedInventoryProduct(t, 4)

	_, err := repositories.AdjustStock(product.ID, 0, "", user.ID)
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}

	var updated models.Product
	database.DB.First(&updated, product.ID)
	if updated.StockQuantity != 0 {
		t.Fatalf("stock = %v want 0", updated.StockQuantity)
	}
}

func TestAdjustStockRejectsNegativeQuantity(t *testing.T) {
	setupInventoryTestDB(t)
	product, user := seedInventoryProduct(t, 10)

	_, err := repositories.AdjustStock(product.ID, -1, "", user.ID)
	if err == nil {
		t.Fatal("expected error for negative quantity")
	}
}

func TestAdjustStockUsesLockedCurrentStockForSecondAdjustment(t *testing.T) {
	setupInventoryTestDB(t)
	product, userA := seedInventoryProduct(t, 20)
	userB := models.User{Username: "worker2", PasswordHash: "hash", Role: models.RoleWorker, IsActive: true}
	if err := database.DB.Create(&userB).Error; err != nil {
		t.Fatalf("create user B: %v", err)
	}

	if _, err := repositories.AdjustStock(product.ID, 18, "A", userA.ID); err != nil {
		t.Fatalf("adjust A: %v", err)
	}

	result, err := repositories.AdjustStock(product.ID, 25, "B", userB.ID)
	if err != nil {
		t.Fatalf("adjust B: %v", err)
	}
	if result.PreviousStock != 18 || result.Change != 7 {
		t.Fatalf("expected change from locked stock 18->25 (+7), got %+v", result)
	}

	var updated models.Product
	if err := database.DB.First(&updated, product.ID).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if updated.StockQuantity != 25 {
		t.Fatalf("final stock = %v want 25", updated.StockQuantity)
	}
}

func TestListInventoryMovementsPaginationAndFilters(t *testing.T) {
	setupInventoryTestDB(t)
	product, user := seedInventoryProduct(t, 10)
	if _, err := repositories.AdjustStock(product.ID, 12, "A", user.ID); err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if _, err := repositories.AdjustStock(product.ID, 8, "B", user.ID); err != nil {
		t.Fatalf("adjust: %v", err)
	}

	all, total, err := repositories.ListInventoryMovements(repositories.MovementListQuery{
		Page:  1,
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(all) != 1 {
		t.Fatalf("pagination: total=%d len=%d", total, len(all))
	}

	filtered, totalFiltered, err := repositories.ListInventoryMovements(repositories.MovementListQuery{
		Page:      1,
		Limit:     20,
		ProductID: "999",
	})
	if err != nil {
		t.Fatalf("filter product: %v", err)
	}
	if totalFiltered != 0 || len(filtered) != 0 {
		t.Fatalf("expected empty filter result")
	}

	byType, totalType, err := repositories.ListInventoryMovements(repositories.MovementListQuery{
		Page:         1,
		Limit:        20,
		MovementType: "adjust",
	})
	if err != nil {
		t.Fatalf("filter type: %v", err)
	}
	if totalType != 2 || len(byType) != 2 {
		t.Fatalf("type filter: total=%d len=%d", totalType, len(byType))
	}
}

func TestLowStockExcludeOutOfStock(t *testing.T) {
	setupInventoryTestDB(t)
	product, _ := seedInventoryProduct(t, 0)

	products, total, err := repositories.ListLowStockProducts(repositories.LowStockQuery{
		Page:  1,
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("list low stock: %v", err)
	}
	if total != 1 || len(products) != 1 {
		t.Fatalf("expected zero stock in low-stock list")
	}

	excluded, totalExcluded, err := repositories.ListLowStockProducts(repositories.LowStockQuery{
		Page:              1,
		Limit:             20,
		ExcludeOutOfStock: true,
	})
	if err != nil {
		t.Fatalf("list excluded: %v", err)
	}
	if totalExcluded != 0 || len(excluded) != 0 {
		t.Fatalf("expected zero stock excluded, got %d", totalExcluded)
	}

	product.StockQuantity = 3
	database.DB.Save(product)

	included, totalIncluded, err := repositories.ListLowStockProducts(repositories.LowStockQuery{
		Page:              1,
		Limit:             20,
		ExcludeOutOfStock: true,
	})
	if err != nil {
		t.Fatalf("list included: %v", err)
	}
	if totalIncluded != 1 || len(included) != 1 {
		t.Fatalf("expected low stock product, got %d", totalIncluded)
	}
}
