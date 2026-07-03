package models

import "gorm.io/gorm"

type Invoice struct {
	gorm.Model

	CreatedByUserID uint
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserID"`

	CustomerID *uint
	Customer   *Customer `gorm:"foreignKey:CustomerID"`

	TotalAmount float64       `gorm:"not null"`
	PaidAmount  float64       `gorm:"not null;default:0"`
	Status      InvoiceStatus `gorm:"not null"`
	Items       []InvoiceItem `gorm:"foreignKey:InvoiceID"`
}

type InvoiceStatus string

const (
	InvoiceStatusPaid   InvoiceStatus = "paid"
	InvoiceStatusUnpaid InvoiceStatus = "unpaid"
	InvoiceStatusPartiallyPaid InvoiceStatus = "partiallyPaid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)	

func IsValidInvoiceStatus(status string) bool {//ako je jedan od ova dva stringa, vrati true
	switch InvoiceStatus(status) {
		case InvoiceStatusPaid, InvoiceStatusUnpaid, InvoiceStatusPartiallyPaid, InvoiceStatusCancelled:
			return true
		default:
			return false
	}
}
//imamo status tipa InvoiceStatus koji je string i ima mogucnosti paid i unpaid
//ovo je funkcija koja provjerava da li je status validan
//u handleru posle na osnovu ove funkcije provjeravamo da li je status validan i ako nije, vracamo error


func IsValidInvoiceSort(status string) bool {
	return status == "createdAt" || 
		status == "totalAmount"
}

func IsValidSortDirection(direction string) bool {
	return direction == "asc" || direction == "desc"
}

