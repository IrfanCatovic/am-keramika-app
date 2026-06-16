package dto

type AddStockRequest struct {

	ProductID uint `json:"productID" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,min=0"`
	Note string `json:"note"`
}

type AdjustStockRequest struct {
	ProductID uint `json:"productID" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,ne=0"`
	Note string `json:"note"`
}

type SellStockRequest struct {
	ProductID uint `json:"productID" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Note string `json:"note"`
}
