package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name string `gorm:"not null"`
	Slug string `gorm:"unique;not null"`
	Description string

	CategoryID int `gorm:"not null"`
	Category Category `gorm:"foreignKey:CategoryID"`

	Unit string `gorm:"not null"`

	PruchasePrice float64 `gorm:"not null"`
	MarginPercent float64 `gorm:"default:0.0"`
	VatPercent float64 `gorm:"default:0.0"`
	SalePrice float64 `gorm:"not null"`

	StockQuantity int `gorm:"default:0"`
	MinStockQuantity int `gorm:"default:0"`

	HasVariants bool `gorm:"default:false"`

	IsActive bool `gorm:"default:true"`
	IsOnSale bool `gorm:"default:false"`
	ShowOnHomepage bool `gorm:"default:false"`
	
}