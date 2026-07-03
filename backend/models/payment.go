package models

import "gorm.io/gorm"

type Payment struct {
	gorm.Model

	CustomerID uint `gorm:"not null"`
	CreatedByUserID uint `gorm:"not null"`
	TotalAmount float64 `gorm:"not null"`

	Allocations []PaymentAllocation `gorm:"foreignKey:PaymentID"`
}

type PaymentAllocation struct {
	gorm.Model

	PaymentID uint `gorm:"not null"`
	InvoiceID uint `gorm:"not null"`
	Amount float64 `gorm:"not null"`

	Payment Payment `gorm:"foreignKey:PaymentID"`
	Invoice Invoice `gorm:"foreignKey:InvoiceID"`
}

