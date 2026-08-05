package models

import "gorm.io/gorm"

type InvoiceCancellation struct {
	gorm.Model
	InvoiceID uint    `gorm:"not null;uniqueIndex"`
	Invoice   Invoice `gorm:"foreignKey:InvoiceID"`

	CreatedByUserID uint
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserID"`

	Reason            string  `gorm:"not null"`
	DebtReducedAmount float64 `gorm:"not null;default:0"`
	RefundedAmount    float64 `gorm:"not null;default:0"`
}
