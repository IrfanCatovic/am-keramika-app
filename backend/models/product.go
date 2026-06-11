package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name string `gorm:"not null"`
	Slug string `gorm:"unique;not null"`
	Description string

	CategoryID uint `gorm:"not null"`
	Category Category `gorm:"foreignKey:CategoryID"`

	Unit string `gorm:"not null"`

	PruchasePrice *float64 
	MarginPercent *float64 
	VatPercent *float64 //zvezdica znaci da je ovo optional moze biti prazno 
	SalePrice float64 `gorm:"not null"`

	StockQuantity float64 `gorm:"default:0"`
	MinStockQuantity float64 `gorm:"default:0"`

	HasVariants bool `gorm:"default:false"`

	IsActive bool `gorm:"default:true"`
	IsOnSale bool `gorm:"default:false"`
	ShowOnHomepage bool `gorm:"default:false"`
	
}