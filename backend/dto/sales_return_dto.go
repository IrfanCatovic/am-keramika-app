package dto

// CreateSalesReturnRequest is the staff payload for a multi-item goods return.
// unitPrice is the explicit return price chosen by the worker (not catalog price).
// cashRefunded must be sent explicitly (true or false); omit is rejected.
type CreateSalesReturnRequest struct {
	Description  string                         `json:"description"`
	CashRefunded *bool                          `json:"cashRefunded"`
	Items        []CreateSalesReturnItemRequest `json:"items" binding:"required"`
}

type CreateSalesReturnItemRequest struct {
	ProductID uint    `json:"productID" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required"`
	UnitPrice float64 `json:"unitPrice" binding:"required"`
}

type SalesReturnItemResponse struct {
	ID              uint    `json:"id"`
	ProductID       uint    `json:"productID"`
	ProductName     string  `json:"productName"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unitPrice"`
	TotalPrice      float64 `json:"totalPrice"`
	SaleByPackage   bool    `json:"saleByPackage"`
	PackageQuantity float64 `json:"packageQuantity"`
}

type SalesReturnResponse struct {
	ID            uint                      `json:"id"`
	Description   string                    `json:"description"`
	TotalAmount   float64                   `json:"totalAmount"`
	CashRefunded  bool                      `json:"cashRefunded"`
	RefundID      *uint                     `json:"refundID,omitempty"`
	CreatedAt     string                    `json:"createdAt"`
	CreatedByUser *UserSummaryResponse      `json:"createdByUser,omitempty"`
	Items         []SalesReturnItemResponse `json:"items"`
}
