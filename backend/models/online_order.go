package models

import "gorm.io/gorm"

type OnlineOrderStatus string

const (
	OnlineOrderStatusPending   OnlineOrderStatus = "pending"
	OnlineOrderStatusConfirmed OnlineOrderStatus = "confirmed" // KORAK 5
)

// OnlineOrder is a public checkout request. It is NOT an Invoice.
// Stock is not reserved or decremented until staff confirms (KORAK 5).
type OnlineOrder struct {
	gorm.Model

	Status OnlineOrderStatus `gorm:"size:32;not null;index;default:pending"`

	FirstName string `gorm:"size:100;not null"`
	LastName  string `gorm:"size:100;not null"`
	Phone     string `gorm:"size:50;not null"`
	City      string `gorm:"size:150;not null"`
	Address   string `gorm:"size:250;not null"`
	Email     string `gorm:"size:254"`
	Note      string `gorm:"size:1000"`

	TotalAmount float64 `gorm:"not null"`

	Items []OnlineOrderItem `gorm:"foreignKey:OnlineOrderID"`
}

type OnlineOrderItem struct {
	gorm.Model

	OnlineOrderID uint        `gorm:"not null;index"`
	OnlineOrder   OnlineOrder `gorm:"foreignKey:OnlineOrderID"`

	ProductID uint    `gorm:"not null;index"`
	Product   Product `gorm:"foreignKey:ProductID"`

	// Snapshots of what the customer ordered (immutable after create).
	ProductName string  `gorm:"size:255;not null"`
	ProductSlug string  `gorm:"size:255"`
	Unit        string  `gorm:"size:50;not null"`
	Quantity    float64 `gorm:"not null"`
	UnitPrice   float64 `gorm:"not null"`
	TotalPrice  float64 `gorm:"not null"`
}
