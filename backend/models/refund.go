package models

import "gorm.io/gorm"

// Refund is a cash outflow record.
// Cancellation refunds keep InvoiceID set (one per invoice).
// Sales-return refunds keep SalesReturnID set and InvoiceID nil.
type Refund struct {
	gorm.Model

	InvoiceID *uint    `gorm:"uniqueIndex"`
	Invoice   *Invoice `gorm:"foreignKey:InvoiceID"`

	SalesReturnID *uint        `gorm:"uniqueIndex"`
	SalesReturn   *SalesReturn `gorm:"foreignKey:SalesReturnID"`

	CreatedByUserID uint
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserID"`

	Amount float64 `gorm:"not null"`
	Reason string  `gorm:"not null"`
}
