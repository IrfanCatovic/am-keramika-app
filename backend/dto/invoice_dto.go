package dto

type CreateInvoiceRequest struct {
	CustomerName string                     `json:"customerName" `
	Items        []CreateInvoiceItemRequest `json:"items" binding:"required"`
}

type CreateInvoiceItemRequest struct {
	ProductID uint `json:"productID" binding:"required"`

	Quantity float64 `json:"quantity" binding:"required,gt=0"`
}

type InvoiceItemResponse struct {
	ProductID   uint    `json:"productID"`
	ProductName string  `json:"productName"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalPrice  float64 `json:"totalPrice"`
}

type InvoiceResponse struct {
	ID           uint                  `json:"id"`
	CustomerName string                `json:"customerName"`
	TotalAmount  float64               `json:"totalAmount"`
	Status       string                `json:"status"`
	Items        []InvoiceItemResponse `json:"items"`
}

type InvoiceListResponse struct {
	ID           uint    `json:"id"`
	CustomerName string  `json:"customerName"`
	TotalAmount  float64 `json:"totalAmount"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
}

type PaginatedInvoiceResponse struct {

	Data []InvoiceListResponse `json:"data"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int64 `json:"total"`
	TotalPages int `json:"totalPages"`
}
