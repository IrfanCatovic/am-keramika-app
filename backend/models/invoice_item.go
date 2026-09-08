package models

import "gorm.io/gorm"

type InvoiceItem struct {
	gorm.Model

	InvoiceID uint    `gorm:"not null"`
	Invoice   Invoice `gorm:"foreignKey:InvoiceID"`

	ProductID uint    `gorm:"not null"`
	Product   Product `gorm:"foreignKey:ProductID"`

	// Quantity is the actual billable/issued quantity. The remaining fields
	// preserve the package contract that was active when the invoice was made.
	Quantity          float64 `gorm:"not null"`
	RequestedQuantity float64 `gorm:"not null;default:0"`
	SaleByPackage     bool    `gorm:"not null;default:false"`
	PackageQuantity   float64 `gorm:"not null;default:0"`
	PackageCount      int     `gorm:"not null;default:0"`
	// OriginalUnitPrice is the catalog effective unit price at sale time
	// (before any cashier override). Nil on legacy rows created before this field.
	OriginalUnitPrice *float64 `gorm:""`
	// UnitPrice is the final billed unit price for this line.
	UnitPrice       float64 `gorm:"not null"`
	PriceOverridden bool    `gorm:"not null;default:false"`
	TotalPrice      float64 `gorm:"not null"`
}
