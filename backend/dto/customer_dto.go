package dto

type CreateCustomerDTO struct {
	Name string `json:"name" binding:"required"`
	Phone string `json:"phone"`
}

