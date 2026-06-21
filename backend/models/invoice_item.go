package models

import "gorm.io/gorm"

type InvoiceItem struct {
	gorm.Model

	InvoiceID uint `gorm:"not null"`
	Invoice Invoice `gorm:"foreignKey:InvoiceID"`

	ProductID uint `gorm:"not null"`
	Product Product `gorm:"foreignKey:ProductID"`

	Quantity float64 `gorm:"not null"`
	UnitPrice float64 `gorm:"not null"`
	TotalPrice float64 `gorm:"not null"`
}