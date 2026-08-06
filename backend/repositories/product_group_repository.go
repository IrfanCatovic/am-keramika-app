package repositories

import (
	"errors"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

func CreateProductGroup(group *models.ProductGroup) error {
	if err := validateCategoryActive(group.CategoryID); err != nil {
		return err
	}

	var existing models.ProductGroup
	err := database.DB.
		Where("category_id = ? AND slug = ?", group.CategoryID, group.Slug).
		First(&existing).Error
	if err == nil {
		return errors.New("grupa sa ovim slugom već postoji u kategoriji")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return database.DB.Create(group).Error
}

func GetAllProductGroups(categoryID string) ([]models.ProductGroup, error) {
	var groups []models.ProductGroup
	query := database.DB.Preload("Category")

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Order("id ASC").Find(&groups).Error
	return groups, err
}

func GetProductGroupByID(id uint) (*models.ProductGroup, error) {
	var group models.ProductGroup
	err := database.DB.Preload("Category").First(&group, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("grupa proizvoda nije pronađena")
		}
		return nil, err
	}
	return &group, nil
}

func UpdateProductGroup(group *models.ProductGroup) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var category models.Category
		if err := tx.First(&category, group.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCategoryNotFound
			}
			return err
		}
		if !category.IsActive {
			return ErrCategoryInactive
		}

		var current models.ProductGroup
		if err := tx.First(&current, group.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("grupa proizvoda nije pronađena")
			}
			return err
		}

		if current.CategoryID != group.CategoryID {
			var productCount int64
			if err := tx.Model(&models.Product{}).Where("group_id = ?", group.ID).Count(&productCount).Error; err != nil {
				return err
			}
			if productCount > 0 {
				return errors.New("grupa ima proizvode; premjestite ili uklonite proizvode iz grupe prije promjene kategorije")
			}
		}

		var existing models.ProductGroup
		err := tx.
			Where("category_id = ? AND slug = ? AND id <> ?", group.CategoryID, group.Slug, group.ID).
			First(&existing).Error
		if err == nil {
			return errors.New("grupa sa ovim slugom već postoji u kategoriji")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return tx.Save(group).Error
	})
}

func DeleteProductGroup(id uint) error {
	var group models.ProductGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("grupa proizvoda nije pronađena")
		}
		return err
	}

	var productCount int64
	if err := database.DB.Model(&models.Product{}).Where("group_id = ?", id).Count(&productCount).Error; err != nil {
		return err
	}
	if productCount > 0 {
		return errors.New("grupa ima proizvode; premjestite ili uklonite proizvode prije brisanja")
	}

	return database.DB.Delete(&group).Error
}
