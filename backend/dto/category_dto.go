package dto

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCategoryStatusRequest struct {
	IsActive bool `json:"isActive"`
}

type CategoryResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
}
