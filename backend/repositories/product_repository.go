package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	DefaultProductListPage  = 1
	DefaultProductListLimit = 20
	MaxProductListLimit     = 100
)

type ProductListQuery struct {
	Search          string
	CategoryID      string
	GroupID         string
	Ungrouped       bool
	IncludeInactive bool
	Page            int
	Limit           int
}

func validateProductGroupAssignment(categoryID uint, groupID *uint) error {
	if groupID == nil {
		return nil
	}

	var group models.ProductGroup
	if err := database.DB.First(&group, *groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("grupa proizvoda nije pronađena")
		}
		return err
	}

	if group.CategoryID != categoryID {
		return errors.New("grupa ne pripada izabranoj kategoriji")
	}

	return nil
}

func buildProductListQuery(q ProductListQuery) *gorm.DB {
	query := database.DB.Model(&models.Product{})

	if !q.IncludeInactive {
		query = query.
			Where("products.is_active = ?", true).
			Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true)
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
	if q.Ungrouped {
		query = query.Where("products.group_id IS NULL")
	}

	return query
}

func ListProducts(q ProductListQuery) ([]models.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultProductListPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultProductListLimit
	}

	var total int64
	if err := buildProductListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []models.Product
	offset := (q.Page - 1) * q.Limit
	err := buildProductListQuery(q).
		Preload("Category").
		Preload("Group").
		Order("products.name ASC, products.id ASC").
		Limit(q.Limit).
		Offset(offset).
		Find(&products).Error
	return products, total, err
}

func CreateProduct(product *models.Product) error {
	if err := validateCategoryActive(product.CategoryID); err != nil {
		return err
	}

	if err := validateProductGroupAssignment(product.CategoryID, product.GroupID); err != nil {
		return err
	}

	return database.DB.Create(product).Error
}

func GetProductById(id string) (*models.Product, error) {
	var product models.Product
	result := database.DB.
		Preload("Category").
		Preload("Group").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, sort_order ASC, id ASC")
		}).
		First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

func UpdateProduct(product *models.Product) error {
	if err := validateCategoryActive(product.CategoryID); err != nil {
		return err
	}

	if err := validateProductGroupAssignment(product.CategoryID, product.GroupID); err != nil {
		return err
	}

	// Explicit Select osigurava da se group_id može postaviti na NULL.
	return database.DB.Model(product).
		Select("Name", "Slug", "Description", "CategoryID", "GroupID", "Unit", "SalePrice", "StockQuantity", "PurchasePrice", "MarginPercent").
		Updates(product).Error
}

func DeactivateProduct(id string) error {
	var productID uint
	if _, err := fmt.Sscanf(id, "%d", &productID); err != nil {
		return errors.New("proizvod nije pronađen")
	}

	hasImages, err := ProductHasImages(productID)
	if err != nil {
		return err
	}
	if hasImages {
		return ErrProductHasImages
	}

	result := database.DB.Model(&models.Product{}).Where("id = ?", id).Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("proizvod nije pronađen")
	}

	return nil
}
