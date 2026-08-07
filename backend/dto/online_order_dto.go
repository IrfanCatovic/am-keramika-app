package dto

// --- Public create (existing) ---

type PublicCreateOnlineOrderRequest struct {
	FirstName string                              `json:"firstName"`
	LastName  string                              `json:"lastName"`
	Phone     string                              `json:"phone"`
	City      string                              `json:"city"`
	Address   string                              `json:"address"`
	Email     string                              `json:"email"`
	Note      string                              `json:"note"`
	Website   string                              `json:"website"`
	Items     []PublicCreateOnlineOrderItemRequest `json:"items"`
}

type PublicCreateOnlineOrderItemRequest struct {
	ProductID uint    `json:"productID"`
	Quantity  float64 `json:"quantity"`
}

type PublicOnlineOrderResponse struct {
	ID          uint    `json:"id"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"totalAmount"`
	CreatedAt   string  `json:"createdAt"`
}

type PublicOnlineOrderErrorResponse struct {
	Message   string `json:"message"`
	ProductID *uint  `json:"productID,omitempty"`
	Code      string `json:"code,omitempty"`
}

// --- Internal staff ---

type OnlineOrderPendingCountResponse struct {
	Count int64 `json:"count"`
}

type OnlineOrderListItemResponse struct {
	ID          uint    `json:"id"`
	Status      string  `json:"status"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Phone       string  `json:"phone"`
	City        string  `json:"city"`
	TotalAmount float64 `json:"totalAmount"`
	ItemsCount  int     `json:"itemsCount"`
	CreatedAt   string  `json:"createdAt"`
	ConfirmedAt *string `json:"confirmedAt,omitempty"`
	InvoiceID   *uint   `json:"invoiceID,omitempty"`
}

type OnlineOrderListResponse struct {
	Orders     []OnlineOrderListItemResponse `json:"orders"`
	Pagination ProductPaginationResponse     `json:"pagination"`
}

type OnlineOrderItemDetailResponse struct {
	ProductID              uint    `json:"productID"`
	ProductName            string  `json:"productName"`
	ProductSlug            string  `json:"productSlug"`
	Unit                   string  `json:"unit"`
	Quantity               float64 `json:"quantity"`
	UnitPrice              float64 `json:"unitPrice"`
	TotalPrice             float64 `json:"totalPrice"`
	CurrentProductActive   *bool   `json:"currentProductActive,omitempty"`
	CurrentInStockEnough   *bool   `json:"currentInStockEnough,omitempty"`
}

type OnlineOrderDetailResponse struct {
	ID          uint                            `json:"id"`
	Status      string                          `json:"status"`
	FirstName   string                          `json:"firstName"`
	LastName    string                          `json:"lastName"`
	Phone       string                          `json:"phone"`
	City        string                          `json:"city"`
	Address     string                          `json:"address"`
	Email       string                          `json:"email"`
	Note        string                          `json:"note"`
	TotalAmount float64                         `json:"totalAmount"`
	CreatedAt   string                          `json:"createdAt"`
	ConfirmedAt *string                         `json:"confirmedAt,omitempty"`
	InvoiceID   *uint                           `json:"invoiceID,omitempty"`
	Items       []OnlineOrderItemDetailResponse `json:"items"`
}

type ConfirmOnlineOrderNewCustomer struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type ConfirmOnlineOrderRequest struct {
	CustomerID  *uint                          `json:"customerID"`
	NewCustomer *ConfirmOnlineOrderNewCustomer `json:"newCustomer"`
}

type ConfirmOnlineOrderResponse struct {
	OrderID   uint `json:"orderID"`
	InvoiceID uint `json:"invoiceID"`
	Status    string `json:"status"`
}

type ConfirmOnlineOrderErrorResponse struct {
	Message   string `json:"message"`
	ProductID *uint  `json:"productID,omitempty"`
	Code      string `json:"code,omitempty"`
}
