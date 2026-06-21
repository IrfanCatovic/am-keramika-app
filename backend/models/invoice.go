package models

import "gorm.io/gorm"

type Invoice struct {
	gorm.Model

	CreatedByUserID uint
	CreatedByUser User `gorm:"foreignKey:CreatedByUserID"`

	TotalAmount float64 `gorm:"not null"`
	Status string `gorm:"not null"`
	Items []InvoiceItem `gorm:"foreignKey:InvoiceID"`


}