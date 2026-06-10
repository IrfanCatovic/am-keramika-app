package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
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