package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model

	Name     string `gorm:"unique;not null"`
	Slug     string `gorm:"unique;not null"`
	IsActive bool   `gorm:"not null;default:true"`

	Products      []Product      `gorm:"foreignKey:CategoryID"`
	ProductGroups []ProductGroup `gorm:"foreignKey:CategoryID"`
}
