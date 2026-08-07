package dto

// Customer invoice paymentMode (optional; empty/"unpaid" = legacy unpaid).
const (
	InvoicePaymentModeUnpaid  = "unpaid"
	InvoicePaymentModeFull    = "full"
	InvoicePaymentModePartial = "partial"
)

type CreateInvoiceRequest struct {
	CustomerID           *uint                      `json:"customerID"`
	Items                []CreateInvoiceItemRequest `json:"items" binding:"required"`
	PaymentMode          string                     `json:"paymentMode"`
	InitialPaymentAmount *float64                   `json:"initialPaymentAmount"`
}

type CreateInvoiceItemRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type InvoiceItemResponse struct {
	ProductID   uint    `json:"productID"`
	ProductName string  `json:"productName"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalPrice  float64 `json:"totalPrice"`
}

type InvoiceResponse struct {
	ID              uint                         `json:"id"`
	CustomerID      *uint                        `json:"customerID"`
	Customer        *CustomerResponse            `json:"customer"`
	TotalAmount     float64                      `json:"totalAmount"`
	PaidAmount      float64                      `json:"paidAmount"`
	RemainingAmount float64                      `json:"remainingAmount"`
	Status          string                       `json:"status"`
	CreatedAt       string                       `json:"createdAt"`
	CreatedByUser   *UserSummaryResponse         `json:"createdByUser,omitempty"`
	Items           []InvoiceItemResponse        `json:"items"`
	Cancellation    *InvoiceCancellationResponse `json:"cancellation,omitempty"`
	Refund          *RefundResponse              `json:"refund,omitempty"`
}

type InvoiceListResponse struct {
	ID              uint                 `json:"id"`
	CustomerID      *uint                `json:"customerID"`
	Customer        *CustomerResponse    `json:"customer"`
	CustomerName    string               `json:"customerName,omitempty"`
	TotalAmount     float64              `json:"totalAmount"`
	PaidAmount      float64              `json:"paidAmount"`
	RemainingAmount float64              `json:"remainingAmount"`
	Status          string               `json:"status"`
	CreatedAt       string               `json:"createdAt"`
	CreatedByUser   *UserSummaryResponse `json:"createdByUser,omitempty"`
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
