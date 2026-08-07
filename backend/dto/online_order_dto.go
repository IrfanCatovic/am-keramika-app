package dto

// PublicCreateOnlineOrderRequest is the public checkout payload.
// Client-sent prices/names are ignored — backend recomputes everything.
type PublicCreateOnlineOrderRequest struct {
	FirstName string                              `json:"firstName"`
	LastName  string                              `json:"lastName"`
	Phone     string                              `json:"phone"`
	City      string                              `json:"city"`
	Address   string                              `json:"address"`
	Email     string                              `json:"email"`
	Note      string                              `json:"note"`
	Website   string                              `json:"website"` // honeypot; must be empty
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
