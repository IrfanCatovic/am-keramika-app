package repositories

import "am-keramika-backend/database"

func CreateCustomer(customer *models.Customer) error {

	result := database.DB.Create(customer) 
	return result.Error
}