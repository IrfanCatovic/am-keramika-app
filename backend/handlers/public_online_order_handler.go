package handlers

import (
	"errors"
	"net/http"
	"time"

	"am-keramika-backend/dto"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func CreatePublicOnlineOrder(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MiB

	var req dto.PublicCreateOnlineOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PublicOnlineOrderErrorResponse{
			Message: "Proverite unesene podatke i pokušajte ponovo.",
			Code:    "validation",
		})
		return
	}

	order, err := repositories.CreateOnlineOrder(req)
	if err != nil {
		var createErr *repositories.OnlineOrderCreateError
		if errors.As(err, &createErr) {
			status := http.StatusBadRequest
			message := createErr.Error()
			switch {
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderHoneypot):
				// Silent success-looking rejection would leak; use 400 without details.
				c.JSON(http.StatusBadRequest, dto.PublicOnlineOrderErrorResponse{
					Message: "Proverite unesene podatke i pokušajte ponovo.",
					Code:    "validation",
				})
				return
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderInsufficientStock):
				message = "Nema dovoljno proizvoda na stanju za izabranu količinu."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderProductUnavailable):
				message = "Jedan od proizvoda više nije dostupan."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderInvalidEmail):
				message = "Email adresa nije ispravna."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderTooManyItems):
				message = "Previše stavki u narudžbini."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderInvalidQuantity):
				message = "Količina mora biti veća od 0."
			default:
				message = "Proverite unesene podatke i pokušajte ponovo."
			}
			c.JSON(status, dto.PublicOnlineOrderErrorResponse{
				Message:   message,
				ProductID: createErr.ProductID,
				Code:      createErr.Code,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.PublicOnlineOrderErrorResponse{
			Message: "Narudžbinu trenutno nije moguće poslati. Pokušajte ponovo.",
			Code:    "server_error",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.PublicOnlineOrderResponse{
		ID:          order.ID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt.UTC().Format(time.RFC3339),
	})

	// Best-effort notification — never fails the customer response.
	go notifyNewOnlineOrder(order)
}
