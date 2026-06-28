package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
)


func CreateCustomer(customer *models.Customer) error {

	result := database.DB.Create(customer) 
	return result.Error
}