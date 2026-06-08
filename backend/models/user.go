package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	FullName string `gorm:"not null"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role string `gorm:"default:'worker'"`

	IsActive bool `gorm:"default:true"`

}