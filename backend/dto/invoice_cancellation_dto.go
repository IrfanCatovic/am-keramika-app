package dto

type CancelInvoiceRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
}