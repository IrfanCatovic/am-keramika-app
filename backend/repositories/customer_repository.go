package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
)


func CreateCustomer(customer *models.Customer) error {

	result := database.DB.Create(customer) 
	return result.Error
}

func GetAllCustomers(page int, limit int) ([]models.Customer, int64, error) {

	var customers []models.Customer
	var total int64

	query := database.DB.Model(&models.Customer{}) //radimo sa customer tabelom
	offset := (page - 1) * limit //offset je broj redova koje preskacemo, racuna se strana 3, limit 20  
	//(3-1) * 20 = 40, preskacemo 40 redova i pocinjemo sa 41. redom
	err := query.Count(&total).Error//brojimo ukupan broj redova u tabeli
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil

}