package dto

type CreateProductRequest struct{
	Name string
	CategoryID uint
	Unit string
	SalePrice float64
	StockQuantity float64
	Description string
}

type UpdateProductRequest struct{
	Name string
	CategoryID uint
	Unit string
	SalePrice float64
	StockQuantity float64
	Description string
}