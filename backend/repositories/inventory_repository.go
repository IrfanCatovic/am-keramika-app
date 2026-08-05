package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
)

type SellStockResult struct {
	Warning string
}

func AddStock(productID uint, quantity float64, note string, createdByUserID uint) error {
	tx := database.DB.Begin() //ovde kreiramo transakciju

	var product models.Product
	err := tx.First(&product, productID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	product.StockQuantity += quantity

	err = tx.Save(&product).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	movement := models.InventoryMovement{
		ProductID:       productID,
		CreatedByUserID: createdByUserID,
		MovementType:    "in",
		Quantity:        quantity,
		Note:            note,
	}

	err = tx.Create(&movement).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func AdjustStock(productID uint, quantity float64, note string, createdByUserID uint) error {
	tx := database.DB.Begin()

	var product models.Product
	err := tx.First(&product, productID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	product.StockQuantity += quantity

	err = tx.Save(&product).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	movement := models.InventoryMovement{
		ProductID:       productID,
		CreatedByUserID: createdByUserID,
		MovementType:    "adjust",
		Quantity:        quantity,
		Note:            note,
	}

	err = tx.Create(&movement).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func SellStock(productID uint, quantity float64, note string, createdByUserID uint) (*SellStockResult, error) {
	tx := database.DB.Begin()

	var product models.Product
	err := tx.First(&product, productID).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if product.StockQuantity < quantity {
		tx.Rollback()
		return nil, errors.New("nema dovoljno robe na stanju")
	}

	product.StockQuantity -= quantity

	err = tx.Save(&product).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	movement := models.InventoryMovement{
		ProductID:       productID,
		CreatedByUserID: createdByUserID,
		MovementType:    "sale",
		Quantity:        quantity,
		Note:            note,
	}

	err = tx.Create(&movement).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	result := &SellStockResult{}

	if product.StockQuantity < product.MinStockQuantity {
		result.Warning = "Proizvod je pao ispod minimalnog lagera"
	}

	return result, nil

}
