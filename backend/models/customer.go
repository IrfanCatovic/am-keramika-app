package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	Name      string `gorm:"not null"`
	Phone     string
	TotalDebt float64 `gorm:"default:0"`
	IsActive  bool    `gorm:"not null;default:true"`

	Invoices []Invoice `gorm:"foreignKey:CustomerID"`
}
