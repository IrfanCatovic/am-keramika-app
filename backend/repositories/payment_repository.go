package repositories

import (
	"errors"
	"fmt"
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
		return models.Payment{}, errors.New("uplata mora imati bar jedan račun")
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

		if invoice.Status == models.InvoiceStatusPaid {
			tx.Rollback()
			return models.Payment{}, fmt.Errorf("racun %d je vec placen", invoice.ID)
		}
		if invoice.Status == models.InvoiceStatusCancelled {
			tx.Rollback()
			return models.Payment{}, fmt.Errorf("ne moze se izvrsiti uplata na storniran racun %d", invoice.ID)
		}

		remainingAmount := invoice.TotalAmount - invoice.PaidAmount

		if allocationReq.Amount > remainingAmount { //provera da li je iznos prenosenja veci od ostatka racuna
			tx.Rollback()
			return models.Payment{}, errors.New("iznos uplate ne može biti veći od preostalog duga računa")
		}

		invoice.PaidAmount += allocationReq.Amount //ovo je lokalna promena, ne utice na bazu, dodajemo iznos prenosenja na iznos placenog racuna	

		//sa frontenda nam stize 300, a duzni smo 300 za taj racun onda se uradi, a nije placeno nista 
		//znaci 300 je dug - 0 (koliko je vec otplaceno) = 300  
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

	payment := models.Payment{
		CustomerID: customer.ID,
		CreatedByUserID: createdByUserID,
		TotalAmount: totalAmount,
	}
	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return models.Payment{}, err
	}

	for _, allocationReq := range req.Allocations {
		allocation := models.PaymentAllocation{
			PaymentID: payment.ID,
			InvoiceID: allocationReq.InvoiceID,
			Amount: allocationReq.Amount,
		}
		allocationsToCreate = append(allocationsToCreate, allocation)
	}

	if err := tx.Create(&allocationsToCreate).Error; err != nil {
		tx.Rollback()
		return models.Payment{}, err
	}

	for _, invoice := range invoicesToUpdate {
		if err := tx.Model(&models.Invoice{}).
		 	Where("id = ?", invoice.ID).
			Updates(map[string]interface{}{
				"paid_amount": invoice.PaidAmount,
				"status": invoice.Status,
			}).Error; err != nil {
			tx.Rollback()
			return models.Payment{}, err
		}
	}
	
	newCustomerDebt := customer.TotalDebt - totalAmount
	if newCustomerDebt < 0 {
		tx.Rollback()
		return models.Payment{}, errors.New("kupac ne moze imati negativan dug")
	}
 
	if err := tx.Model(&models.Customer{}).
		Where("id = ?", customer.ID).
		Update("total_debt", newCustomerDebt).Error; err != nil {
		tx.Rollback()
		return models.Payment{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return models.Payment{}, err
	}

	var createdPayment models.Payment
	if err := database.DB.Preload("Customer").
	Preload("CreatedByUser").
	Preload("Allocations").
	Preload("Allocations.Invoice").
	First(&createdPayment, payment.ID).Error; err != nil {
		return payment, nil
	}
	return createdPayment, nil
}