package repositories

import (
	"errors"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/database"

	"gorm.io/gorm"
)

func CreatePayment(req dto.CreatePaymentRequest, createdByUserID uint) (models.Payment, error){
	tx := database.DB.Begin() //zapocinjemo transakciju

	if tx.Error != nil {
		return models.Payment{}, tx.Error
	}

	var customer models.Customer		
	if err:= tx.First(&customer, req.CustomerID).Error; err != nil {
		tx.Rollback()
		
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Payment{}, errors.New("kupac ne postoji")
		}
		return models.Payment{}, err
	}


	tx.Rollback()

	return models.Payment{}, errors.New("create payment nije jos implementiran")
}