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

	//Provera da li postoji slice sa raspodelom racuna i da li postoji duplikat raspodele racuna
	if len(req.Allocations) == 0 {
		tx.Rollback()
		return models.Payment{}, errors.New("Uplata mora imati bar jedna racun")
	}

	seenInvoiceIDs := make(map[uint]bool)
	for  _, allocationReq := range req.Allocations { 
		if allocationReq.Amount <= 0 {
			tx.Rollback()
			return models.Payment{}, errors.New("iznos alokacije mora biti pozitivan")
		}
		if seenInvoiceIDs[allocationReq.InvoiceID] {
			tx.Rollback()
			return models.Payment{}, errors.New("isti račun ne može biti dodat dva puta u jednoj uplati")
		}

		seenInvoiceIDs[allocationReq.InvoiceID] = true
	}

	tx.Rollback()

	return models.Payment{}, errors.New("create payment nije jos implementiran")
}