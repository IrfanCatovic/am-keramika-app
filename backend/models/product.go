package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model

	Name string

	Slug string `gorm:"unique;not null"`

	Description string

	CategoryID uint
	Category Category `gorm:"foreignKey:CategoryID"`

	Unit string

	PurchasePrice *float64
	MarginPercent *float64
	VatPercent *float64

	SalePrice float64

	StockQuantity float64
	MinStockQuantity float64 `gorm:"default:0"`

	HasVariants bool `gorm:"default:false"`

	IsActive bool `gorm:"default:true"`
	IsOnSale bool `gorm:"default:false"`
	ShowOnHomepage bool `gorm:"default:false"`
}