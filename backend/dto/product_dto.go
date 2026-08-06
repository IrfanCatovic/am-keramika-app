package dto

type CreateProductRequest struct {
	Name             string   `json:"name" binding:"required"`
	CategoryID       uint     `json:"categoryID" binding:"required"`
	GroupID          *uint    `json:"groupID"`
	Unit             string   `json:"unit" binding:"required"`
	SalePrice        float64  `json:"salePrice" binding:"required,gt=0"`
	StockQuantity    float64  `json:"stockQuantity" binding:"gte=0"`
	MinStockQuantity float64  `json:"minStockQuantity" binding:"gte=0"`
	Description      string   `json:"description"`
	PurchasePrice    *float64 `json:"purchasePrice"`
	MarginPercent    *float64 `json:"marginPercent"`
}

type UpdateProductRequest struct {
	Name             string       `json:"name" binding:"required"`
	CategoryID       uint         `json:"categoryID" binding:"required"`
	GroupID          OptionalUint `json:"groupID"`
	Unit             string       `json:"unit" binding:"required"`
	SalePrice        float64      `json:"salePrice" binding:"required,gt=0"`
	StockQuantity    float64      `json:"stockQuantity" binding:"gte=0"`
	MinStockQuantity float64      `json:"minStockQuantity" binding:"gte=0"`
	Description      string       `json:"description"`
	PurchasePrice    *float64     `json:"purchasePrice"`
	MarginPercent    *float64     `json:"marginPercent"`
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
	ID               uint                            `json:"id"`
	Name             string                          `json:"name"`
	Slug             string                          `json:"slug"`
	Description      string                          `json:"description"`
	CategoryID       uint                            `json:"categoryID"`
	Category         *ProductCategorySummaryResponse `json:"category,omitempty"`
	GroupID          *uint                           `json:"groupID"`
	Group            *ProductGroupSummaryResponse    `json:"group,omitempty"`
	Unit             string                          `json:"unit"`
	SalePrice        float64                         `json:"salePrice"`
	StockQuantity    float64                         `json:"stockQuantity"`
	MinStockQuantity float64                         `json:"minStockQuantity"`
	IsActive         bool                            `json:"isActive"`
	PurchasePrice    *float64                        `json:"purchasePrice,omitempty"`
	MarginPercent    *float64                        `json:"marginPercent,omitempty"`
	Images           []ProductImageResponse          `json:"images,omitempty"`
	PrimaryImage     *ProductImageResponse           `json:"primaryImage"`
}

type ProductPaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedProductListResponse struct {
	Products   []ProductResponse         `json:"products"`
	Pagination ProductPaginationResponse `json:"pagination"`
}
