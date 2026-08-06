package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type SellStockResult struct {
	Warning string
}

const (
	DefaultLowStockPage  = 1
	DefaultLowStockLimit = 20
	MaxLowStockLimit     = 100
)

type LowStockQuery struct {
	Page       int
	Limit      int
	Search     string
	CategoryID string
	GroupID    string
}

func buildLowStockQuery(q LowStockQuery) *gorm.DB {
	query := database.DB.Model(&models.Product{}).
		Where("products.is_active = ?", true).
		Where("products.stock_quantity <= products.min_stock_quantity").
		Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true)

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

func AddStock(productID uint, quantity float64, note string, createdByUserID uint) error {
	tx := database.DB.Begin() //ovde kreiramo transakciju

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

func AdjustStock(productID uint, quantity float64, note string, createdByUserID uint) error {
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
		MovementType:    "adjust",
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
