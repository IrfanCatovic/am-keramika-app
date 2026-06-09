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