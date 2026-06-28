package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	Name string `gorm:"not null"`
	Phone string
	IsActive bool `gorm:"default:true"`

	Invoices []Invoice `gorm:"foreignKey:CustomerID"`
}