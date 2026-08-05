package models

import "gorm.io/gorm"

type Refund struct {
	gorm.Model

	InvoiceID uint    `gorm:"not null;uniqueIndex"`
	Invoice   Invoice `gorm:"foreignKey:InvoiceID"`

	CreatedByUserID uint
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserID"`

	Amount float64 `gorm:"not null"`
	Reason string  `gorm:"not null"`
}
