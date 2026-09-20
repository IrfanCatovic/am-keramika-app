package models

import "gorm.io/gorm"

// SalesReturn represents one customer return visit with one or more products.
// It does not depend on an Invoice and does not create Refund / stock side effects by itself.
type SalesReturn struct {
	gorm.Model

	TotalAmount float64 `gorm:"not null"`
	Description string  `gorm:"size:1000"`

	CreatedByUserID uint
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserID"`

	Items []SalesReturnItem `gorm:"foreignKey:SalesReturnID"`
}

// SalesReturnItem is one returned product line inside a SalesReturn.
// ProductName, Unit, SaleByPackage and PackageQuantity are immutable snapshots.
// Quantity is the exact physical quantity received — never rounded up to packages.
type SalesReturnItem struct {
	gorm.Model

	SalesReturnID uint        `gorm:"not null;index"`
	SalesReturn   SalesReturn `gorm:"foreignKey:SalesReturnID"`

	ProductID uint    `gorm:"not null;index"`
	Product   Product `gorm:"foreignKey:ProductID"`

	ProductName string `gorm:"size:255;not null"`
	Unit        string `gorm:"size:50;not null"`

	Quantity   float64 `gorm:"not null"`
	UnitPrice  float64 `gorm:"not null"`
	TotalPrice float64 `gorm:"not null"`

	SaleByPackage   bool    `gorm:"not null;default:false"`
	PackageQuantity float64 `gorm:"not null;default:0"`
}
