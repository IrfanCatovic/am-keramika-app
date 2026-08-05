package models

import "gorm.io/gorm"

type ProductGroup struct {
	gorm.Model
	Name string
	Slug string `gorm:"not null;uniqueIndex:idx_product_group_category_slug"`
	CategoryID uint `gorm:"not null;index;uniqueIndex:idx_product_group_category_slug"`

	Category Category `gorm:"foreignKey:CategoryID"`
	Products []Product `gorm:"foreignKey:ProductGroupID"`
}

