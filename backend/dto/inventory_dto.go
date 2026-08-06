package dto

type AddStockRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,min=0"`
	Note      string  `json:"note"`
}

type AdjustStockRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,ne=0"`
	Note      string  `json:"note"`
}

type SellStockRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Note      string  `json:"note"`
}

type LowStockCategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type LowStockGroupResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type LowStockProductResponse struct {
	ID               uint                     `json:"id"`
	Name             string                   `json:"name"`
	Unit             string                   `json:"unit"`
	StockQuantity    float64                  `json:"stockQuantity"`
	MinStockQuantity float64                  `json:"minStockQuantity"`
	MissingQuantity  float64                  `json:"missingQuantity"`
	Category         *LowStockCategoryResponse `json:"category"`
	Group            *LowStockGroupResponse    `json:"group"`
	PrimaryImage     *ProductImageResponse     `json:"primaryImage"`
}

type LowStockPaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedLowStockResponse struct {
	Products   []LowStockProductResponse  `json:"products"`
	Pagination LowStockPaginationResponse `json:"pagination"`
}
