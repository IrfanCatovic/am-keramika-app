package dto

type CreatePaymentRequest struct {
	CustomerID  uint                             `json:"customerID,omitempty" binding:"required"`
	TotalAmount float64                          `json:"totalAmount" binding:"required,gt=0"`
	Allocations []CreatePaymentAllocationRequest `json:"allocations" binding:"required,min=1"`
}

type CreatePaymentAllocationRequest struct {
	InvoiceID uint    `json:"invoiceID" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentCustomerResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	IsActive bool   `json:"isActive"`
}

type PaymentUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

type PaymentResponse struct {
	ID              uint                        `json:"id"`
	CustomerID      *uint                       `json:"customerID,omitempty"`
	Customer        *PaymentCustomerResponse    `json:"customer,omitempty"`
	CreatedByUserID uint                        `json:"createdByUserID"`
	CreatedByUser   *PaymentUserResponse        `json:"createdByUser,omitempty"`
	TotalAmount     float64                     `json:"totalAmount"`
	CreatedAt       string                      `json:"createdAt"`
	Allocations     []PaymentAllocationResponse `json:"allocations"`
}

type PaymentAllocationResponse struct {
	ID        uint                             `json:"id"`
	InvoiceID uint                             `json:"invoiceID"`
	Amount    float64                          `json:"amount"`
	Invoice   PaymentAllocationInvoiceResponse `json:"invoice"`
}

type PaymentAllocationInvoiceResponse struct {
	ID          uint    `json:"id"`
	TotalAmount float64 `json:"totalAmount"`
	PaidAmount  float64 `json:"paidAmount"`
	Status      string  `json:"status"`
}

type PaginatedPaymentResponse struct {
	Data       []PaymentResponse `json:"data"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
}
