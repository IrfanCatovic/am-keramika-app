package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	FullName     string `json:"fullName"`
	Username     string `gorm:"unique;not null" json:"username"`
	PasswordHash string `gorm:"column:password;not null" json:"-"`
	Role         string `gorm:"not null;default:'radnik'" json:"role"`
	IsActive     bool   `gorm:"default:true" json:"isActive"`
}
