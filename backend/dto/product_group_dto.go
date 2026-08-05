package dto

type CreateProductGroupRequest struct {
	Name       string `json:"name" binding:"required"`
	CategoryID uint   `json:"categoryID" binding:"required"`
	Slug       string `json:"slug"`
}

type UpdateProductGroupRequest struct {
	Name       string `json:"name" binding:"required"`
	CategoryID uint   `json:"categoryID" binding:"required"`
	Slug       string `json:"slug"`
}

type ProductGroupResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	CategoryID uint   `json:"categoryID"`
}

type ProductGroupListResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	CategoryID uint   `json:"categoryID"`
}
