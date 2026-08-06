package repositories

import (
	"errors"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

var (
	ErrCategoryDuplicateName        = errors.New("kategorija sa ovim nazivom već postoji")
	ErrCategoryDuplicateSlug        = errors.New("kategorija sa ovim slugom već postoji")
	ErrCategoryNotFound             = errors.New("kategorija nije pronađena")
	ErrCategoryInactive             = errors.New("kategorija nije aktivna")
	ErrCategoryHasGroupsOrProducts  = errors.New("kategorija ima grupe ili proizvode; deaktivirajte kategoriju umjesto brisanja")
)

func validateCategoryActive(categoryID uint) error {
	var category models.Category
	if err := database.DB.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	if !category.IsActive {
		return ErrCategoryInactive
	}
	return nil
}

func categoryNameExists(name string, excludeID uint) (bool, error) {
	var existing models.Category
	query := database.DB.Where("name = ?", name)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.First(&existing).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func categorySlugExists(slug string, excludeID uint) (bool, error) {
	var existing models.Category
	query := database.DB.Where("slug = ?", slug)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.First(&existing).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func CreateCategory(category *models.Category) error {
	exists, err := categoryNameExists(category.Name, 0)
	if err != nil {
		return err
	}
	if exists {
		return ErrCategoryDuplicateName
	}

	exists, err = categorySlugExists(category.Slug, 0)
	if err != nil {
		return err
	}
	if exists {
		return ErrCategoryDuplicateSlug
	}

	category.IsActive = true
	return database.DB.Create(category).Error
}

func GetCategories(includeInactive bool) ([]models.Category, error) {
	var categories []models.Category
	query := database.DB
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("id ASC").Find(&categories).Error
	return categories, err
}

func GetCategoryByID(id string) (*models.Category, error) {
	var category models.Category
	result := database.DB.First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, result.Error
	}
	return &category, nil
}

func UpdateCategory(category *models.Category) error {
	exists, err := categoryNameExists(category.Name, category.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrCategoryDuplicateName
	}

	exists, err = categorySlugExists(category.Slug, category.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrCategoryDuplicateSlug
	}

	return database.DB.Model(category).Select("Name", "Slug").Updates(category).Error
}

func UpdateCategoryStatus(id uint, isActive bool) error {
	result := database.DB.Model(&models.Category{}).Where("id = ?", id).Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func DeleteCategory(id uint) error {
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}

	var groupCount int64
	if err := database.DB.Model(&models.ProductGroup{}).Where("category_id = ?", id).Count(&groupCount).Error; err != nil {
		return err
	}

	var productCount int64
	if err := database.DB.Model(&models.Product{}).Where("category_id = ?", id).Count(&productCount).Error; err != nil {
		return err
	}

	if groupCount > 0 || productCount > 0 {
		return ErrCategoryHasGroupsOrProducts
	}

	return database.DB.Delete(&category).Error
}
