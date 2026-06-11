package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
)

func CreateProduct(product *models.Product) error {
	result := database.DB.Create(&product)

	return result.Error	
}

func GetAllProducts() ([]models.Product, error) {
  var products []models.Product

  result := database.DB.Preload("Category").Find(&products)

  return products, result.Error
}

func GetProductById(id string) (*models.Product, error) {
	var product models.Product
	result := database.DB.Preload("Category").First(&product, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &product, nil 
}

func UpdateProduct(product *models.Product) error {
	result := database.DB.Save(&product)
	return result.Error
}

func DeactivateProduct(id string) error {
	result := database.DB.Model(&models.Product{}).Where("id = ?", id).Update("active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("proizvod nije pronađen")
	}

	return nil 
}