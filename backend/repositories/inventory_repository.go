package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SellStockResult struct {
	Warning string
}

type AdjustStockResult struct {
	ProductID     uint
	PreviousStock float64
	NewStock      float64
	Change        float64
	MovementID    uint
}

const (
	DefaultLowStockPage       = 1
	DefaultLowStockLimit      = 20
	MaxLowStockLimit          = 100
	DefaultMovementListPage   = 1
	DefaultMovementListLimit  = 20
	MaxMovementListLimit      = 100
)

type LowStockQuery struct {
	Page               int
	Limit              int
	Search             string
	CategoryID         string
	GroupID            string
	ExcludeOutOfStock  bool
}

type MovementListQuery struct {
	Page         int
	Limit        int
	ProductID    string
	MovementType string
	FromDate     string
	ToDate       string
}

func buildLowStockQuery(q LowStockQuery) *gorm.DB {
	query := database.DB.Model(&models.Product{}).
		Where("products.is_active = ?", true).
		Where("products.stock_quantity <= products.min_stock_quantity").
		Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true)

	if q.ExcludeOutOfStock {
		query = query.Where("products.stock_quantity > 0")
	}

	if q.Search != "" {
		search := strings.ToLower(strings.TrimSpace(q.Search))
		pattern := "%" + search + "%"
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.slug) LIKE ?", pattern, pattern)
	}
	if q.CategoryID != "" {
		query = query.Where("products.category_id = ?", q.CategoryID)
	}
	if q.GroupID != "" {
		query = query.Where("products.group_id = ?", q.GroupID)
	}

	return query
}

func ListLowStockProducts(q LowStockQuery) ([]models.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultLowStockPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultLowStockLimit
	}

	var total int64
	if err := buildLowStockQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []models.Product
	offset := (q.Page - 1) * q.Limit
	err := buildLowStockQuery(q).
		Preload("Category").
		Preload("Group").
		Order("(products.stock_quantity - products.min_stock_quantity) ASC, products.name ASC, products.id ASC").
		Limit(q.Limit).
		Offset(offset).
		Find(&products).Error
	return products, total, err
}

func CountLowStockProducts(excludeOutOfStock bool) (int64, error) {
	var total int64
	err := buildLowStockQuery(LowStockQuery{ExcludeOutOfStock: excludeOutOfStock}).Count(&total).Error
	return total, err
}

func CountOutOfStockProducts() (int64, error) {
	var total int64
	err := database.DB.Model(&models.Product{}).
		Where("products.is_active = ?", true).
		Where("products.stock_quantity <= 0").
		Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true).
		Count(&total).Error
	return total, err
}

func buildMovementListQuery(q MovementListQuery) *gorm.DB {
	query := database.DB.Model(&models.InventoryMovement{})

	if q.ProductID != "" {
		query = query.Where("inventory_movements.product_id = ?", q.ProductID)
	}
	if movementType := strings.TrimSpace(q.MovementType); movementType != "" {
		query = query.Where("inventory_movements.movement_type = ?", movementType)
	}
	if q.FromDate != "" {
		if parsed, err := time.Parse("2006-01-02", q.FromDate); err == nil {
			query = query.Where("inventory_movements.created_at >= ?", parsed)
		}
	}
	if q.ToDate != "" {
		if parsed, err := time.Parse("2006-01-02", q.ToDate); err == nil {
			end := parsed.Add(24 * time.Hour)
			query = query.Where("inventory_movements.created_at < ?", end)
		}
	}

	return query
}

func ListInventoryMovements(q MovementListQuery) ([]models.InventoryMovement, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultMovementListPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultMovementListLimit
	}

	var total int64
	if err := buildMovementListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var movements []models.InventoryMovement
	offset := (q.Page - 1) * q.Limit
	err := buildMovementListQuery(q).
		Preload("Product").
		Preload("CreatedByUser").
		Order("inventory_movements.created_at DESC, inventory_movements.id DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&movements).Error
	return movements, total, err
}

func AddStock(productID uint, quantity float64, note string, createdByUserID uint) error {
	tx := database.DB.Begin()

	var product models.Product
	err := tx.First(&product, productID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	product.StockQuantity += quantity

	err = tx.Save(&product).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	movement := models.InventoryMovement{
		ProductID:       productID,
		CreatedByUserID: createdByUserID,
		MovementType:    "in",
		Quantity:        quantity,
		Note:            note,
	}

	err = tx.Create(&movement).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func AdjustStock(productID uint, newQuantity float64, note string, createdByUserID uint) (*AdjustStockResult, error) {
	if newQuantity < 0 {
		return nil, errors.New("nova količina ne sme biti negativna")
	}

	tx := database.DB.Begin()

	var product models.Product
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("proizvod nije pronađen")
		}
		return nil, err
	}

	if !product.IsActive {
		tx.Rollback()
		return nil, errors.New("proizvod nije aktivan")
	}

	previous := product.StockQuantity
	change := newQuantity - previous

	if change != 0 {
		product.StockQuantity = newQuantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		movement := models.InventoryMovement{
			ProductID:       productID,
			CreatedByUserID: createdByUserID,
			MovementType:    "adjust",
			Quantity:        change,
			Note:            note,
		}

		if err := tx.Create(&movement).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		return &AdjustStockResult{
			ProductID:     productID,
			PreviousStock: previous,
			NewStock:      newQuantity,
			Change:        change,
			MovementID:    movement.ID,
		}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &AdjustStockResult{
		ProductID:     productID,
		PreviousStock: previous,
		NewStock:      newQuantity,
		Change:        0,
		MovementID:    0,
	}, nil
}

func SellStock(productID uint, quantity float64, note string, createdByUserID uint) (*SellStockResult, error) {
	tx := database.DB.Begin()

	var product models.Product
	err := tx.First(&product, productID).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if product.StockQuantity < quantity {
		tx.Rollback()
		return nil, errors.New("nema dovoljno robe na stanju")
	}

	product.StockQuantity -= quantity

	err = tx.Save(&product).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	movement := models.InventoryMovement{
		ProductID:       productID,
		CreatedByUserID: createdByUserID,
		MovementType:    "sale",
		Quantity:        quantity,
		Note:            note,
	}

	err = tx.Create(&movement).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	result := &SellStockResult{}

	if product.StockQuantity < product.MinStockQuantity {
		result.Warning = "Proizvod je pao ispod minimalnog lagera"
	}

	return result, nil

}
