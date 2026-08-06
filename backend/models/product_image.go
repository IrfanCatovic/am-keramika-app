package models

import "gorm.io/gorm"

type ProductImage struct {
	gorm.Model

	ProductID uint    `gorm:"not null;index"`
	Product   Product `gorm:"foreignKey:ProductID"`

	URL       string `gorm:"not null"`
	PublicID  string `gorm:"unique;not null"`
	IsPrimary bool   `gorm:"not null;default:false"`
	SortOrder int    `gorm:"not null;default:0"`

	Width  *int
	Height *int
	Format string
	Bytes  *int64
}
