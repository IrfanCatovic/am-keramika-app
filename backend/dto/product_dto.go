package dto

type CreateProductRequest struct {
	Name             string   `json:"name" binding:"required"`
	CategoryID       uint     `json:"categoryID" binding:"required"`
	GroupID          *uint    `json:"groupID"`
	Unit             string   `json:"unit" binding:"required"`
	SalePrice        *float64 `json:"salePrice"`
	StockQuantity    float64  `json:"stockQuantity" binding:"gte=0"`
	MinStockQuantity float64  `json:"minStockQuantity" binding:"gte=0"`
	Description      string   `json:"description"`
	PurchasePrice    *float64 `json:"purchasePrice"`
	MarginPercent    *float64 `json:"marginPercent"`
	VatPercent       *float64 `json:"vatPercent"`
	IsOnSale         bool     `json:"isOnSale"`
	DiscountPercent  *float64 `json:"discountPercent"`
	ShowOnHomepage   bool     `json:"showOnHomepage"`
}

type UpdateProductRequest struct {
	Name             string       `json:"name" binding:"required"`
	CategoryID       uint         `json:"categoryID" binding:"required"`
	GroupID          OptionalUint `json:"groupID"`
	Unit             string       `json:"unit" binding:"required"`
	SalePrice        *float64     `json:"salePrice"`
	StockQuantity    float64      `json:"stockQuantity" binding:"gte=0"`
	MinStockQuantity float64      `json:"minStockQuantity" binding:"gte=0"`
	Description      string       `json:"description"`
	PurchasePrice    *float64     `json:"purchasePrice"`
	MarginPercent    *float64     `json:"marginPercent"`
	VatPercent       *float64     `json:"vatPercent"`
	IsActive         *bool        `json:"isActive"`
	IsOnSale         *bool        `json:"isOnSale"`
	DiscountPercent  *float64     `json:"discountPercent"`
	ShowOnHomepage   *bool        `json:"showOnHomepage"`
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
	SalePrice          float64                         `json:"salePrice"`
	EffectiveSalePrice float64                         `json:"effectiveSalePrice"`
	StockQuantity      float64                         `json:"stockQuantity"`
	MinStockQuantity   float64                         `json:"minStockQuantity"`
	IsActive           bool                            `json:"isActive"`
	IsOnSale           bool                            `json:"isOnSale"`
	DiscountPercent    float64                         `json:"discountPercent"`
	ShowOnHomepage     bool                            `json:"showOnHomepage"`
	PricingMode        string                          `json:"pricingMode"`
	PurchasePrice      *float64                        `json:"purchasePrice,omitempty"`
	MarginPercent      *float64                        `json:"marginPercent,omitempty"`
	VatPercent         *float64                        `json:"vatPercent,omitempty"`
	Images             []ProductImageResponse          `json:"images,omitempty"`
	PrimaryImage       *ProductImageResponse           `json:"primaryImage"`
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
