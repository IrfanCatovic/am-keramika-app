package repositories

import (
	"errors"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"gorm.io/gorm"
)

func CreateInvoice(req dto.CreateInvoiceRequest, createdByUserID uint) (*models.Invoice, error) {
	tx := database.DB.Begin()

	if req.CustomerID != nil {
		var customer models.Customer

		err := tx.First(&customer, *req.CustomerID).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.New("kupac nije pronađen")
		}
	}

	invoiceStatus := models.InvoiceStatusPaid
	if req.CustomerID != nil {
		invoiceStatus = models.InvoiceStatusUnpaid
	}

	invoice := models.Invoice{
		CreatedByUserID: createdByUserID,
		CustomerID:      req.CustomerID,
		Status:          invoiceStatus,
		TotalAmount:     0,
		PaidAmount:      0,
	}

	err := tx.Create(&invoice).Error //kreira fakturu u bazi da bi dobili ID fakture
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var totalAmount float64 = 0

	for _, item := range req.Items {
		var product models.Product

		err := tx.First(&product, item.ProductID).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.New("proizvod nije pronađen")
		}

		if product.StockQuantity < item.Quantity {
			tx.Rollback()
			return nil, errors.New("nema dovoljno stoka na skladištu")
		}

		unitPrice := product.SalePrice
		totalPrice := unitPrice * item.Quantity

		invoiceItem := models.InvoiceItem{
			InvoiceID:  invoice.ID,
			ProductID:  product.ID,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: totalPrice,
		}

		err = tx.Create(&invoiceItem).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		product.StockQuantity -= item.Quantity

		err = tx.Save(&product).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		//ovaj blok radimo ako nema kupca odmah cim se napravi racun pravi se i placanje i raspodela placanja na racun
		if req.CustomerID == nil {
			payment := models.Payment{
				CustomerID:      nil,
				CreatedByUserID: createdByUserID,
				TotalAmount:     totalAmount,
			}

			err = tx.Create(&payment).Error
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			allocation := models.PaymentAllocation{
				PaymentID: payment.ID,
				InvoiceID: invoice.ID,
				Amount:    totalAmount,
			}
			err = tx.Create(&allocation).Error
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			movement := models.InventoryMovement{
				ProductID:       product.ID,
				CreatedByUserID: createdByUserID,
				MovementType:    "sale",
				Quantity:        item.Quantity,
				Note:            "Prodaja kroz racun",
			}

			err = tx.Create(&movement).Error
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			totalAmount += totalPrice
		}

		invoice.TotalAmount = totalAmount

		if req.CustomerID == nil {
			invoice.PaidAmount = totalAmount
			invoice.Status = models.InvoiceStatusPaid
		} else {
			invoice.PaidAmount = 0
			invoice.Status = models.InvoiceStatusUnpaid

			err = tx.Model(&models.Customer{}).Where("id = ?", *req.CustomerID).Update("total_debt", gorm.Expr("total_debt + ?", totalAmount)).Error

			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		err = tx.Save(&invoice).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		err = tx.Commit().Error
		if err != nil {
			return nil, err
		}

		err = database.DB.Preload("Items").Preload("Items.Product").First(&invoice, invoice.ID).Error
		if err != nil {
			return nil, err
		}
	}
	return &invoice, nil
}

func GetInvoiceByID(id uint) (*models.Invoice, error) {

	var invoice models.Invoice

	err := database.DB.
		Preload("Customer").
		Preload("CreatedByUser").
		Preload("Items").
		Preload("Items.Product").
		First(&invoice, id).Error

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func GetAllInvoices(page int, limit int, search string, status string, sort string, direction string) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := database.DB.Model(&models.Invoice{}) //Radicemo sa invoice tabelom

	if search != "" {
		query = query.Joins("JOIN customers ON customers.id = invoices.customer_id"). //radimo JOIN tj INNER JOIN sa customers tabelom, jer se fakture spajaju sa kupcima preko customer_id
												Where("customers.name ILIKE ?", "%"+search+"%") //Inner join ne vraca nule ako nema kupca, dok left join vraca nule ako nema kupca
	}

	if status != "" {
		query = query.Where("invoices.status = ?", status)
	}

	sortColumn := "created_at"
	sortDirection := "DESC"

	if sort == "totalAmount" {
		sortColumn = "total_amount"
	}
	if direction == "asc" {
		sortDirection = "ASC"
	}

	offset := (page - 1) * limit

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Customer").Order(sortColumn + " " + sortDirection).Limit(limit).Offset(offset).Find(&invoices).Error
	if err != nil {
		return nil, 0, err
	}
	return invoices, total, nil
}

func GetOpenInvoicesByCustomerID(customerID uint) ([]models.Invoice, error) {
	var customer models.Customer

	err := database.DB.First(&customer, customerID).Error
	if err != nil {
		return nil, errors.New("kupac nije pronađen")
	}

	var invoices []models.Invoice

	err = database.DB.Where("customer_id = ? AND status IN ?", customerID, []models.InvoiceStatus{
		models.InvoiceStatusUnpaid,
		models.InvoiceStatusPartiallyPaid,
	}).Order("created_at DESC").Find(&invoices).Error

	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func CancelInvoice(id uint, req dto.CancelInvoiceRequest, createdByUserID uint) (*dto.CancelInvoiceResponse, error) {
	var invoice models.Invoice

	tx := database.DB.Begin()

	err := tx.Model(&models.Invoice{}).Preload("Items").First(&invoice, id).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if invoice.Status == models.InvoiceStatusCancelled {
		tx.Rollback()
		return nil, errors.New("racun je vec otkazan")
	}

	remainingAmount := invoice.TotalAmount - invoice.PaidAmount
	refundedAmount := invoice.PaidAmount

	for _, item := range invoice.Items {
		var product models.Product
		err = tx.First(&product, item.ProductID).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		product.StockQuantity += item.Quantity
		err = tx.Save(&product).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		inventoryMovement := models.InventoryMovement{
			ProductID:       item.ProductID,
			CreatedByUserID: createdByUserID,
			MovementType:    "return",
			Quantity:        item.Quantity,
			Note:            "Otkazan racun",
		}
		err = tx.Create(&inventoryMovement).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if invoice.CustomerID != nil {
		var customer models.Customer
		err = tx.First(&customer, *invoice.CustomerID).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		customer.TotalDebt -= remainingAmount
		err = tx.Save(&customer).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	var refundID uint
	if refundedAmount > 0 {
		refund := models.Refund{
			InvoiceID:       invoice.ID,
			CreatedByUserID: createdByUserID,
			Amount:          refundedAmount,
			Reason:          req.Reason,
		}
		err = tx.Create(&refund).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		refundID = refund.ID
	}

	invoiceCancellation := models.InvoiceCancellation{
		InvoiceID:         invoice.ID,
		CreatedByUserID:   createdByUserID,
		Reason:            req.Reason,
		DebtReducedAmount: remainingAmount,
		RefundedAmount:    refundedAmount,
	}
	err = tx.Create(&invoiceCancellation).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	invoice.Status = models.InvoiceStatusCancelled
	err = tx.Save(&invoice).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	err = database.DB.Preload("CreatedByUser").First(&invoiceCancellation, invoiceCancellation.ID).Error
	if err != nil {
		return nil, err
	}

	response := &dto.CancelInvoiceResponse{
		ID:                invoiceCancellation.ID,
		InvoiceID:         invoiceCancellation.InvoiceID,
		Reason:            invoiceCancellation.Reason,
		DebtReducedAmount: invoiceCancellation.DebtReducedAmount,
		RefundedAmount:    invoiceCancellation.RefundedAmount,
		CreatedByUser: &dto.UserSummaryResponse{
			ID:       invoiceCancellation.CreatedByUser.ID,
			Username: invoiceCancellation.CreatedByUser.Username,
		},
	}

	if refundID > 0 {
		var refund models.Refund
		err = database.DB.Preload("CreatedByUser").First(&refund, refundID).Error
		if err != nil {
			return nil, err
		}
		response.Refund = &dto.RefundResponse{
			ID:        refund.ID,
			InvoiceID: refund.InvoiceID,
			Amount:    refund.Amount,
			Reason:    refund.Reason,
			CreatedByUser: &dto.UserSummaryResponse{
				ID:       refund.CreatedByUser.ID,
				Username: refund.CreatedByUser.Username,
			},
		}
	}

	return response, nil
}
