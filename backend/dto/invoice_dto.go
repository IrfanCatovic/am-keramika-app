package dto

type CreateInvoiceRequest struct {
	CustomerID *uint                      `json:"customerID"`
	Items      []CreateInvoiceItemRequest `json:"items" binding:"required"`
}

type CreateInvoiceItemRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type InvoiceItemResponse struct {
	ProductID   uint    `json:"productID"`
	ProductName string  `json:"productName"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalPrice  float64 `json:"totalPrice"`
}

type InvoiceResponse struct {
	ID       uint              `json:"id"`
	Customer *CustomerResponse `json:"customer,omitempty"` // Ovde se koristi CustomerResponse
	// ID uint `json:"id"`
	// Name string `json:"name"`
	// Phone string `json:"phone"`
	TotalAmount float64               `json:"totalAmount"`
	Status      string                `json:"status"`
	Items       []InvoiceItemResponse `json:"items"`
}

type InvoiceListResponse struct {
	ID              uint    `json:"id"`
	CustomerName    string  `json:"customerName"`
	TotalAmount     float64 `json:"totalAmount"`
	PaidAmount      float64 `json:"paidAmount"`
	RemainingAmount float64 `json:"remainingAmount"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"createdAt"`
}

type PaginatedInvoiceResponse struct {
	Data       []InvoiceListResponse `json:"data"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"totalPages"`
}

type CustomerOpenInvoicesResponse struct {
	ID              uint    `json:"id"`
	TotalAmount     float64 `json:"totalAmount"`
	PaidAmount      float64 `json:"paidAmount"`
	RemainingAmount float64 `json:"remainingAmount"`
	CreatedAt       string  `json:"createdAt"`
}
