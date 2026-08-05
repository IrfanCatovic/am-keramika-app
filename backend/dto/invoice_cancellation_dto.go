package dto

type CancelInvoiceRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
}

type CancelInvoiceResponse struct {
	ID                uint                 `json:"id"`
	InvoiceID         uint                 `json:"invoiceID"`
	Reason            string               `json:"reason"`
	DebtReducedAmount float64              `json:"debtReducedAmount"`
	RefundedAmount    float64              `json:"refundedAmount"`
	CreatedByUser     *UserSummaryResponse `json:"createdByUser,omitempty"`
	Refund            *RefundResponse      `json:"refund,omitempty"`
}

type RefundResponse struct {
	ID            uint                 `json:"id"`
	InvoiceID     uint                 `json:"invoiceID"`
	Amount        float64              `json:"amount"`
	Reason        string               `json:"reason"`
	CreatedByUser *UserSummaryResponse `json:"createdByUser,omitempty"`
}
