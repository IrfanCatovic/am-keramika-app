package dto

type CreateCustomerRequest struct {
	Name string `json:"name" binding:"required"`
	Phone string `json:"phone"`
}

