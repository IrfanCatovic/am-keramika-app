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

	totalAmount := 0.0
	invoicesToUpdate := []models.Invoice{} //racuni koje cemo posle azurirati
	allocationsToCreate := []models.PaymentAllocation{} //alokacije koje cemo posle kreirati

	for _, allocationReq := range req.Allocations {
		var invoice models.Invoice

		if err := tx.First(&invoice, allocationReq.InvoiceID).Error; err != nil {
			tx.Rollback()

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.Payment{}, errors.New("racun ne postoji")
			}
			return models.Payment{}, err
		}

		if invoice.CustomerID == nil || *invoice.CustomerID != customer.ID {
			tx.Rollback()
			return models.Payment{}, errors.New("racun ne pripada kupcu")
		}

		remainingAmount := invoice.TotalAmount - invoice.PaidAmount

		if allocationReq.Amount > remainingAmount {
			tx.Rollback()
			return models.Payment{}, errors.New("iznos prenosenja ne moze biti veci od ostatka racuna")
		}

		invoice.PaidAmount += allocationReq.Amount //ovo je lokalna promena, ne utice na bazu

		//sa frontenda nam stize 300, a duzni smo 300 za taj racun onda se uradi, a nije placeno nista 
		//preostala vrednost je 0 i proverimo da li je to sto nam je stiglo sa frontenda vece od preostalog racuna, naravno da nije jer u jednaki ili preostalo vise
		//invoice.paidAmount dodajemo mi ovo sad sto nam je stiglo sa frontenda i proveravamo status ako je invoice.PaindAmount jednak invoice.TotalAmount 
		//onda znaci da je placeno i totalni racun isto tako da je placen, ako nije isto onda je delimicno placen
		if invoice.PaidAmount == invoice.TotalAmount {
			invoice.Status = models.InvoiceStatusPaid
		} else{
			invoice.Status = models.InvoiceStatusPartiallyPaid
		}

		totalAmount += allocationReq.Amount //iznos za payment koji pravimo
		invoicesToUpdate = append(invoicesToUpdate, invoice) //racuni koje cemo posle da azuriramo da ih ne bi cuvali jedan po jedan
	}



	tx.Rollback()

	return models.Payment{}, errors.New("create payment nije jos implementiran")
}