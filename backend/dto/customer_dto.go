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

type PaginatedCustomerResponse struct {
	Data []CustomerListResponse `json:"data"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int64 `json:"total"`
	TotalPages int `json:"total_pages"`
}
