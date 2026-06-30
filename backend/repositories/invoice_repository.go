package repositories

import (
	"errors"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
)

func CreateInvoice(req dto.CreateInvoiceRequest, createdByUserID uint) (*models.Invoice, error) {
	tx := database.DB.Begin()

	invoice := models.Invoice{
		CreatedByUserID: createdByUserID,
		CustomerID: 	 req.CustomerID,
		Status:          "paid",
		TotalAmount:     0,
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

	return &invoice, nil
}

func GetInvoiceByID(id uint) (*models.Invoice, error) {

	var invoice models.Invoice

	err := database.DB.Preload("Customer").Preload("Items").Preload("Items.Product").First(&invoice, id).Error

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func GetAllInvoices(page int, limit int, search string) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64
		
	query := database.DB.Model(&models.Invoice{}) //Radicemo sa invoice tabelom

	if search != "" {
		query = query.Joins("JOIN customers ON customers.id = invoices.customer_id"). //radimo JOIN tj INNER JOIN sa customers tabelom, jer se fakture spajaju sa kupcima preko customer_id
		Where("customers.name LIKE ?", "%"+search+"%") //Inner join ne vraca nule ako nema kupca, dok left join vraca nule ako nema kupca
	}

	offset := (page - 1) * limit

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = query.Preload("Customer").Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices).Error
	if err != nil {
		return nil, 0, err
	}
	return invoices, total, nil
}
	