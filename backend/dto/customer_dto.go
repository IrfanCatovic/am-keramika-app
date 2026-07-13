package dto

type CreateCustomerRequest struct {
	Name string `json:"name" binding:"required"`
	Phone string `json:"phone"`
}

type CustomerResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
}

type CustomerListResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
}

type CustomerDetailsResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
	Debt float64 `json:"debt"`
	Invoices []CustomerInvoiceResponse `json:"invoices"`
}

type CustomerInvoiceResponse struct {
	ID uint `json:"id"`
	TotalAmount float64 `json:"totalAmount"`
	Status string `json:"status"`
	CreatedAt string `json:"createdAt"`
}


type PaginatedCustomerResponse struct {
	Data []CustomerListResponse `json:"data"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int64 `json:"total"`
	TotalPages int `json:"total_pages"`
}

type CustomerFinancialSummaryResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
	TotalDebt float64 `json:"totalDebt"`
	OpenInvoicesCount int64 `json:"openInvoicesCount"`
	PaymentsCount int64 `json:"paymentsCount"`
}

