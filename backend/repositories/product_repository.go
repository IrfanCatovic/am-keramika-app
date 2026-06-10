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

