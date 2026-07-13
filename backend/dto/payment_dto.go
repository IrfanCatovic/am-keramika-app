package dto

type CreatePaymentRequest struct {
	CustomerID uint `json:"customerID" binding:"required"`
	TotalAmount float64 `json:"totalAmount" binding:"required,gt=0"`
	Allocations []CreatePaymentAllocationRequest `json:"allocations" binding:"required,min=1"`
	}

type CreatePaymentAllocationRequest struct {
	InvoiceID uint `json:"invoiceID" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentResponse struct {
	ID uint `json:"id"`
	CustomerID uint `json:"customerID"`
	CreatedByUserID uint `json:"createdByUserID"`
	TotalAmount float64 `json:"totalAmount"`
	CreatedAt string `json:"createdAt"`
	Allocations []PaymentAllocationResponse `json:"allocations"`
}

type PaymentAllocationResponse struct {
	ID uint `json:"id"`
	InvoiceID uint `json:"invoiceID"`
	Amount float64 `json:"amount"`
	Invoice PaymentAllocationInvoiceResponse `json:"invoice"`
}

type PaymentAllocationInvoiceResponse struct {
	ID uint `json:"id"`
	TotalAmount float64 `json:"totalAmount"`
	PaidAmount float64 `json:"paidAmount"`
	Status string `json:"status"`
}