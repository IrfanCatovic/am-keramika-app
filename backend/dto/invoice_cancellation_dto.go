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
	CreatedAt     string               `json:"createdAt,omitempty"`
	CreatedByUser *UserSummaryResponse `json:"createdByUser,omitempty"`
}

type InvoiceCancellationResponse struct {
	ID                uint                 `json:"id"`
	InvoiceID         uint                 `json:"invoiceID"`
	Reason            string               `json:"reason"`
	DebtReducedAmount float64              `json:"debtReducedAmount"`
	RefundedAmount    float64              `json:"refundedAmount"`
	CreatedAt         string               `json:"createdAt"`
	CreatedByUser     *UserSummaryResponse `json:"createdByUser,omitempty"`
}

type RefundListItemResponse struct {
	ID            uint                 `json:"id"`
	InvoiceID     uint                 `json:"invoiceID"`
	Amount        float64              `json:"amount"`
	Reason        string               `json:"reason"`
	CreatedAt     string               `json:"createdAt"`
	CustomerID    *uint                `json:"customerID,omitempty"`
	CustomerName  *string              `json:"customerName,omitempty"`
	CreatedByUser *UserSummaryResponse `json:"createdByUser,omitempty"`
}

type RefundPaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedRefundsResponse struct {
	Refunds    []RefundListItemResponse `json:"refunds"`
	Pagination RefundPaginationResponse `json:"pagination"`
}
