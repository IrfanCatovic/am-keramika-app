package dto

type CreatePaymentRequest struct {
	CustomerID uint `json:"customerID" binding:"required"`
	Allocations []CreatePaymentAllocationRequest `json:"allocations" binding:"required,min=1"`
	}

type CreatePaymentAllocationRequest struct {
	InvoiceID uint `json:"invoiceID" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}