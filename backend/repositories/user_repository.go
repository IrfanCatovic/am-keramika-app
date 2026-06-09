package repositories

import (
	"am-keramika-backend/models"
	"am-keramika-backend/database"
)

func GetAllUsers() ([]models.User, error) {
	var users []models.User
	database.DB.Find(&users)
	return users, nil
}

func GetUserByID(id int)(models.User, error) {
	var user models.User

	result := database.DB.Where("id = ?", id).First(&user) //Where je SQL query za filtriranje podataka u bazi podataka
	if result.Error != nil {
		return models.User{}, result.Error
	}
	return user, nil
}

func GetUserByUsername(username string)(models.User, error) {
	var user models.User
	result := database.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return models.User{}, result.Error
	}
	return user, nil
}

func CreateUser(user *models.User) error {
	result := database.DB.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return result.Error
}