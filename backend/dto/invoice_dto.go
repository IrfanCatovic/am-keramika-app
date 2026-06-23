package dto

type CreateInvoiceRequest struct {
	Items []CreateInvoiceItemRequest `json:"items" binding:"required"`
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
	ID          uint                  `json:"id"`
	TotalAmount float64               `json:"totalAmount"`
	Status      string                `json:"status"`
	Items       []InvoiceItemResponse `json:"items"`
}
