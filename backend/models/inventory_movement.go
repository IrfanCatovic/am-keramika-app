package models

import "gorm.io/gorm"

type InventoryMovement struct {
	gorm.Model

	ProductID    uint
	Product      Product `gorm:"foreignKey:ProductID"`
	MovementType string
	Quantity     float64
	Note         string

	CreatedByUserID uint
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserID"`
}
