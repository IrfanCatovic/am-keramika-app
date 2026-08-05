package dto

type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	CategoryID    uint    `json:"categoryID" binding:"required"`
	GroupID       *uint   `json:"groupID"`
	Unit          string  `json:"unit" binding:"required"`
	SalePrice     float64 `json:"salePrice" binding:"required,gt=0"`
	StockQuantity float64 `json:"stockQuantity" binding:"gte=0"`
	Description   string  `json:"description"`
}

type UpdateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	CategoryID    uint    `json:"categoryID" binding:"required"`
	GroupID       *uint   `json:"groupID"`
	Unit          string  `json:"unit" binding:"required"`
	SalePrice     float64 `json:"salePrice" binding:"required,gt=0"`
	StockQuantity float64 `json:"stockQuantity" binding:"gte=0"`
	Description   string  `json:"description"`
}

type ProductGroupSummaryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProductCategorySummaryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProductResponse struct {
	ID            uint                            `json:"id"`
	Name          string                          `json:"name"`
	Slug          string                          `json:"slug"`
	Description   string                          `json:"description"`
	CategoryID    uint                            `json:"categoryID"`
	Category      *ProductCategorySummaryResponse `json:"category,omitempty"`
	GroupID       *uint                           `json:"groupID"`
	Group         *ProductGroupSummaryResponse    `json:"group,omitempty"`
	Unit          string                          `json:"unit"`
	SalePrice     float64                         `json:"salePrice"`
	StockQuantity float64                         `json:"stockQuantity"`
	IsActive      bool                            `json:"isActive"`
}
