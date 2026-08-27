package models

import (
	"time"

	"gorm.io/gorm"
)

type OnlineOrderStatus string

const (
	OnlineOrderStatusPending   OnlineOrderStatus = "pending"
	OnlineOrderStatusConfirmed OnlineOrderStatus = "confirmed"
)

// OnlineOrder is a public checkout request. It is NOT an Invoice.
// Stock is not reserved or decremented until staff confirms.
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

	InvoiceID *uint    `gorm:"index"`
	Invoice   *Invoice `gorm:"foreignKey:InvoiceID"`

	ConfirmedAt       *time.Time
	ConfirmedByUserID *uint
	ConfirmedByUser   *User `gorm:"foreignKey:ConfirmedByUserID"`

	Items []OnlineOrderItem `gorm:"foreignKey:OnlineOrderID"`
}

type OnlineOrderItem struct {
	gorm.Model

	OnlineOrderID uint        `gorm:"not null;index"`
	OnlineOrder   OnlineOrder `gorm:"foreignKey:OnlineOrderID"`

	ProductID uint    `gorm:"not null;index"`
	Product   Product `gorm:"foreignKey:ProductID"`

	// Snapshots of what the customer ordered (immutable after create).
	ProductName string `gorm:"size:255;not null"`
	ProductSlug string `gorm:"size:255"`
	Unit        string `gorm:"size:50;not null"`
	// Quantity is the actual billable quantity. These fields are immutable
	// snapshots so confirmation never depends on a later Product change.
	Quantity          float64 `gorm:"not null"`
	RequestedQuantity float64 `gorm:"not null;default:0"`
	SaleByPackage     bool    `gorm:"not null;default:false"`
	PackageQuantity   float64 `gorm:"not null;default:0"`
	PackageCount      int     `gorm:"not null;default:0"`
	UnitPrice         float64 `gorm:"not null"`
	TotalPrice        float64 `gorm:"not null"`
}

func IsValidOnlineOrderStatus(status string) bool {
	switch OnlineOrderStatus(status) {
	case OnlineOrderStatusPending, OnlineOrderStatusConfirmed:
		return true
	default:
		return false
	}
}
