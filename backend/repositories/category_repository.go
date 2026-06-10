package repositories

import ("am-keramika-backend/models"
	"am-keramika-backend/database")


func CreateCategory(category models.Category) error {
	result := database.DB.Create(&category)
	return result.Error
}

func GetCategories() ([]models.Category, error) {
	var categories []models.Category
	result := database.DB.Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

