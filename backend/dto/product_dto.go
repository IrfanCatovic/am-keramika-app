package dto

type CreateProductRequest struct {
	Name string `json:"name" binding:"required"`
	CategoryID uint `json:"categoryID" binding:"required"`
	Unit string `json:"unit" binding:"required"`
	SalePrice float64 `json:"salePrice" binding:"required,gt=0"`
	StockQuantity float64 `json:"stockQuantity" binding:"gte=0"`
	Description string `json:"description"`
}
type UpdateProductRequest struct{
	Name string
	CategoryID uint
	Unit string
	SalePrice float64
	StockQuantity float64
	Description string
}