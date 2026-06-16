package dto

type AddStockRequest struct {

	ProductID uint `json:"productID" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,min=0"`
	Note string `json:"note"`
}