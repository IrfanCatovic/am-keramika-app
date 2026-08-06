package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

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

func CreateProduct(product *models.Product) error {
	if err := validateCategoryActive(product.CategoryID); err != nil {
		return err
	}

	if err := validateProductGroupAssignment(product.CategoryID, product.GroupID); err != nil {
		return err
	}

	return database.DB.Create(product).Error
}

func GetAllProducts(search string, categoryID string, includeInactive bool) ([]models.Product, error) {
	var products []models.Product
	query := database.DB.Preload("Category").Preload("Group")

	if !includeInactive {
		query = query.
			Where("products.is_active = ?", true).
			Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true)
	}

	if search != "" {
		query = query.Where("products.name ILIKE ?", "%"+search+"%")
	}
	if categoryID != "" {
		query = query.Where("products.category_id = ?", categoryID)
	}
	result := query.Find(&products)
	return products, result.Error
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
