package dto

type CreateInvoiceRequest struct {
	Items []CreateInvoiceItemRequest `json:"items" binding:"required"`
}

type CreateInvoiceItemRequest struct {
	ProductID uint `json:"productID" binding:"required"`

	Quantity float64 `json:"quantity" binding:"required,gt=0"`

}


