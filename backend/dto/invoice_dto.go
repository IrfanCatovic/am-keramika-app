package dto

type CreateInvoiceItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`

	Quantity float64 `json:"quantity" binding:"required,gt=0"`

}

type CreateInvoiceRequest struct {
	Items []CreateInvoiceItemRequest `json:"items" binding:"required"`
}

