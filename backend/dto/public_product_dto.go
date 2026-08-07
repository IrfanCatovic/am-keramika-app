package dto

// PublicProductResponse is the catalog-facing product shape.
// Never includes purchasePrice, marginPercent, vatPercent, or stockQuantity.
type PublicProductResponse struct {
	ID                 uint                            `json:"id"`
	Name               string                          `json:"name"`
	Slug               string                          `json:"slug"`
	Description        string                          `json:"description"`
	Category           *ProductCategorySummaryResponse `json:"category,omitempty"`
	Group              *ProductGroupSummaryResponse    `json:"group,omitempty"`
	Unit               string                          `json:"unit"`
	SalePrice          float64                         `json:"salePrice"`
	EffectiveSalePrice float64                         `json:"effectiveSalePrice"`
	IsOnSale           bool                            `json:"isOnSale"`
	DiscountPercent    float64                         `json:"discountPercent"`
	InStock            bool                            `json:"inStock"`
	ShowOnHomepage     bool                            `json:"showOnHomepage"`
	Images             []ProductImageResponse          `json:"images,omitempty"`
	PrimaryImage       *ProductImageResponse           `json:"primaryImage"`
}

type PublicCategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PublicProductGroupResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	CategoryID uint   `json:"categoryID"`
}

type PaginatedPublicProductListResponse struct {
	Products   []PublicProductResponse   `json:"products"`
	Pagination ProductPaginationResponse `json:"pagination"`
}

// PublicAvailabilityCheckRequest is the body for POST /public/products/:id/check-availability.
type PublicAvailabilityCheckRequest struct {
	Quantity float64 `json:"quantity"`
}

// PublicAvailabilityCheckResponse never reveals stockQuantity or remaining units.
type PublicAvailabilityCheckResponse struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}
