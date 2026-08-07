package dto

type AddStockRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,min=0"`
	Note      string  `json:"note"`
}

type AdjustStockRequest struct {
	ProductID   uint     `json:"productID" binding:"required"`
	NewQuantity *float64 `json:"newQuantity" binding:"required,min=0"`
	Note        string   `json:"note"`
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

type AdjustStockResponse struct {
	ProductID       uint    `json:"productID"`
	PreviousStock   float64 `json:"previousStock"`
	NewStock        float64 `json:"newStock"`
	Change          float64 `json:"change"`
	MovementID      uint    `json:"movementID"`
}

type InventoryMovementUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

type InventoryMovementResponse struct {
	ID              uint                           `json:"id"`
	ProductID       uint                           `json:"productID"`
	ProductName     string                         `json:"productName"`
	ProductUnit     string                         `json:"productUnit"`
	MovementType    string                         `json:"type"`
	Quantity        float64                        `json:"quantity"`
	Note            string                         `json:"note,omitempty"`
	CreatedAt       string                         `json:"createdAt"`
	CreatedByUser   *InventoryMovementUserResponse `json:"createdByUser,omitempty"`
}

type InventoryMovementPaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedInventoryMovementsResponse struct {
	Movements  []InventoryMovementResponse         `json:"movements"`
	Pagination InventoryMovementPaginationResponse `json:"pagination"`
}
