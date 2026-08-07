package handlers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/mailer"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

var orderMailer mailer.Mailer = mailer.NoopMailer{}

// SetOrderMailer injects the mailer (production SMTP or test double).
func SetOrderMailer(m mailer.Mailer) {
	if m == nil {
		orderMailer = mailer.NoopMailer{}
		return
	}
	orderMailer = m
}

func formatRFC3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatRFC3339Ptr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := formatRFC3339(*t)
	return &s
}

func mapOnlineOrderListItem(order models.OnlineOrder) dto.OnlineOrderListItemResponse {
	return dto.OnlineOrderListItemResponse{
		ID:          order.ID,
		Status:      string(order.Status),
		FirstName:   order.FirstName,
		LastName:    order.LastName,
		Phone:       order.Phone,
		City:        order.City,
		TotalAmount: order.TotalAmount,
		ItemsCount:  len(order.Items),
		CreatedAt:   formatRFC3339(order.CreatedAt),
		ConfirmedAt: formatRFC3339Ptr(order.ConfirmedAt),
		InvoiceID:   order.InvoiceID,
	}
}

func mapOnlineOrderDetail(order models.OnlineOrder) dto.OnlineOrderDetailResponse {
	items := make([]dto.OnlineOrderItemDetailResponse, 0, len(order.Items))
	for _, item := range order.Items {
		row := dto.OnlineOrderItemDetailResponse{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductSlug:  item.ProductSlug,
			Unit:         item.Unit,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			TotalPrice:   item.TotalPrice,
		}
		if order.Status == models.OnlineOrderStatusPending {
			var product models.Product
			err := database.DB.Preload("Category").First(&product, item.ProductID).Error
			if err == nil {
				active := product.IsActive && product.Category.ID != 0 && product.Category.IsActive
				enough := product.StockQuantity >= item.Quantity
				row.CurrentProductActive = &active
				row.CurrentInStockEnough = &enough
			} else {
				f := false
				row.CurrentProductActive = &f
				row.CurrentInStockEnough = &f
			}
		}
		items = append(items, row)
	}

	return dto.OnlineOrderDetailResponse{
		ID:          order.ID,
		Status:      string(order.Status),
		FirstName:   order.FirstName,
		LastName:    order.LastName,
		Phone:       order.Phone,
		City:        order.City,
		Address:     order.Address,
		Email:       order.Email,
		Note:        order.Note,
		TotalAmount: order.TotalAmount,
		CreatedAt:   formatRFC3339(order.CreatedAt),
		ConfirmedAt: formatRFC3339Ptr(order.ConfirmedAt),
		InvoiceID:   order.InvoiceID,
		Items:       items,
	}
}

func GetOnlineOrdersPendingCount(c *gin.Context) {
	count, err := repositories.CountPendingOnlineOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju narudžbina"})
		return
	}
	c.JSON(http.StatusOK, dto.OnlineOrderPendingCountResponse{Count: count})
}

func GetOnlineOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := strings.TrimSpace(c.Query("status"))
	if status != "" && !models.IsValidOnlineOrderStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan status"})
		return
	}

	var fromDate, toDate *time.Time
	if v := strings.TrimSpace(c.Query("fromDate")); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "fromDate mora biti u formatu YYYY-MM-DD"})
			return
		}
		fromDate = &t
	}
	if v := strings.TrimSpace(c.Query("toDate")); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "toDate mora biti u formatu YYYY-MM-DD"})
			return
		}
		end := t.Add(24 * time.Hour)
		toDate = &end
	}

	orders, total, err := repositories.ListOnlineOrders(repositories.OnlineOrderListQuery{
		Page:     page,
		Limit:    limit,
		Status:   status,
		Search:   strings.TrimSpace(c.Query("search")),
		FromDate: fromDate,
		ToDate:   toDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju narudžbina"})
		return
	}

	items := make([]dto.OnlineOrderListItemResponse, 0, len(orders))
	for _, order := range orders {
		items = append(items, mapOnlineOrderListItem(order))
	}

	totalPages := 0
	if total > 0 && limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.OnlineOrderListResponse{
		Orders: items,
		Pagination: dto.ProductPaginationResponse{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

func GetOnlineOrderByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID"})
		return
	}
	order, err := repositories.GetOnlineOrderByID(uint(id))
	if err != nil {
		if errors.Is(err, repositories.ErrOnlineOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Narudžbina nije pronađena"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju narudžbine"})
		return
	}
	c.JSON(http.StatusOK, mapOnlineOrderDetail(*order))
}

func ConfirmOnlineOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID"})
		return
	}

	var req dto.ConfirmOnlineOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ConfirmOnlineOrderErrorResponse{
			Message: "Proverite unesene podatke i pokušajte ponovo.",
			Code:    "validation",
		})
		return
	}

	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Niste prijavljeni"})
		return
	}

	_, invoice, err := repositories.ConfirmOnlineOrder(uint(id), req, userID)
	if err != nil {
		var createErr *repositories.OnlineOrderCreateError
		if errors.As(err, &createErr) {
			status := http.StatusBadRequest
			message := createErr.Error()
			switch {
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderAlreadyProcessed):
				status = http.StatusConflict
				message = "Narudžbina je već obrađena."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderNotFound):
				status = http.StatusNotFound
				message = "Narudžbina nije pronađena."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderConfirmStock):
				message = "Nema dovoljno proizvoda na stanju za potvrdu ove narudžbine."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderConfirmUnavailable):
				message = "Jedan od proizvoda više nije dostupan."
			case errors.Is(createErr.Err, repositories.ErrCustomerInactive),
				errors.Is(createErr.Err, repositories.ErrCustomerNotFound):
				message = "Kupac nije dostupan. Izaberite drugog kupca."
			case errors.Is(createErr.Err, repositories.ErrOnlineOrderConfirmCustomer):
				message = "Izaberite postojećeg kupca ili kreirajte novog."
			default:
				message = "Potvrda narudžbine nije uspela. Pokušajte ponovo."
			}
			c.JSON(status, dto.ConfirmOnlineOrderErrorResponse{
				Message:   message,
				ProductID: createErr.ProductID,
				Code:      createErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ConfirmOnlineOrderErrorResponse{
			Message: "Potvrda narudžbine nije uspela. Pokušajte ponovo.",
			Code:    "server_error",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ConfirmOnlineOrderResponse{
		OrderID:   uint(id),
		InvoiceID: invoice.ID,
		Status:    string(models.OnlineOrderStatusConfirmed),
	})
}

func DeleteOnlineOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID"})
		return
	}

	err = repositories.DeletePendingOnlineOrder(uint(id))
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrOnlineOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": "Narudžbina nije pronađena"})
		case errors.Is(err, repositories.ErrOnlineOrderDeleteNotPending):
			c.JSON(http.StatusConflict, gin.H{"message": "Potvrđena narudžbina se ne može obrisati"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Brisanje nije uspelo"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func notifyNewOnlineOrder(order *models.OnlineOrder) {
	to := mailer.OrderNotificationEmail()
	if to == "" {
		log.Printf("online order #%d: ORDER_NOTIFICATION_EMAIL nije podešen — email nije poslat", order.ID)
		return
	}
	if !orderMailer.Configured() {
		log.Printf("online order #%d: SMTP nije konfigurisan — email nije poslat", order.ID)
		return
	}

	subject := fmt.Sprintf("Nova online narudžbina #%d", order.ID)
	var b strings.Builder
	b.WriteString("Nova online narudžbina\n\n")
	b.WriteString(fmt.Sprintf("Broj: #%d\n\n", order.ID))
	b.WriteString(fmt.Sprintf("Kupac:\n%s %s\n\n", order.FirstName, order.LastName))
	b.WriteString(fmt.Sprintf("Telefon:\n%s\n\n", order.Phone))
	b.WriteString(fmt.Sprintf("Grad:\n%s\n\n", order.City))
	b.WriteString(fmt.Sprintf("Vrednost proizvoda:\n%.2f RSD\n\n", order.TotalAmount))
	b.WriteString(fmt.Sprintf("Stavki:\n%d\n\n", len(order.Items)))
	b.WriteString("Otvorite AM Keramika aplikaciju da pregledate narudžbinu.\n")
	if base := mailer.FrontendAppURL(); base != "" {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s/orders/%d\n", base, order.ID))
	}

	if err := orderMailer.Send(to, subject, b.String()); err != nil {
		log.Printf("online order #%d: slanje emaila nije uspelo: %v", order.ID, err)
	}
}
