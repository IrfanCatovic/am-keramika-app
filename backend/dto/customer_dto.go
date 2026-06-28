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
