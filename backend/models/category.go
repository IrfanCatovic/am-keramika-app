package models

type Category struct {

	gorm.Model
	Name string `gorm:"unique;not null"`
	Slug string `gorm:"unique;not null"`
	IsActive bool `gorm:"default:true"`
}