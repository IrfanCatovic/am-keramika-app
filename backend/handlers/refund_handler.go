package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func mapRefundResponse(refund models.Refund) dto.RefundResponse {
	response := dto.RefundResponse{
		ID:        refund.ID,
		InvoiceID: refund.InvoiceID,
		Amount:    refund.Amount,
		Reason:    refund.Reason,
		CreatedAt: refund.CreatedAt.Format("2006-01-02 15:04"),
	}
	if refund.CreatedByUser.ID != 0 {
		response.CreatedByUser = mapUserSummary(refund.CreatedByUser)
	}
	return response
}

func mapInvoiceCancellationResponse(cancellation models.InvoiceCancellation) dto.InvoiceCancellationResponse {
	response := dto.InvoiceCancellationResponse{
		ID:                cancellation.ID,
		InvoiceID:         cancellation.InvoiceID,
		Reason:            cancellation.Reason,
		DebtReducedAmount: cancellation.DebtReducedAmount,
		RefundedAmount:    cancellation.RefundedAmount,
		CreatedAt:         cancellation.CreatedAt.Format("2006-01-02 15:04"),
	}
	if cancellation.CreatedByUser.ID != 0 {
		response.CreatedByUser = mapUserSummary(cancellation.CreatedByUser)
	}
	return response
}

func mapRefundListItem(refund models.Refund) dto.RefundListItemResponse {
	item := dto.RefundListItemResponse{
		ID:        refund.ID,
		InvoiceID: refund.InvoiceID,
		Amount:    refund.Amount,
		Reason:    refund.Reason,
		CreatedAt: refund.CreatedAt.Format("2006-01-02 15:04"),
	}
	if refund.CreatedByUser.ID != 0 {
		item.CreatedByUser = mapUserSummary(refund.CreatedByUser)
	}
	if refund.Invoice.CustomerID != nil {
		item.CustomerID = refund.Invoice.CustomerID
	}
	if refund.Invoice.Customer != nil {
		name := refund.Invoice.Customer.Name
		item.CustomerName = &name
	}
	return item
}

func enrichInvoiceCancellationData(response *dto.InvoiceResponse, invoiceID uint, status string) {
	if status != string(models.InvoiceStatusCancelled) {
		return
	}

	if cancellation, err := repositories.GetInvoiceCancellationByInvoiceID(invoiceID); err == nil {
		mapped := mapInvoiceCancellationResponse(*cancellation)
		response.Cancellation = &mapped
	}

	if refund, err := repositories.GetRefundByInvoiceID(invoiceID); err == nil {
		mapped := mapRefundResponse(*refund)
		response.Refund = &mapped
	}
}

func GetRefunds(c *gin.Context) {
	if _, err := auth.GetRole(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	page := repositories.DefaultRefundListPage
	if pageStr := c.Query("page"); pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "page mora biti pozitivan broj"})
			return
		}
		page = parsed
	}

	limit := repositories.DefaultRefundListLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit mora biti pozitivan broj"})
			return
		}
		if parsed > repositories.MaxRefundListLimit {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit ne smije biti veći od 100"})
			return
		}
		limit = parsed
	}

	location, err := time.LoadLocation("Europe/Belgrade")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju vremenske zone"})
		return
	}

	query := repositories.RefundListQuery{
		Page:       page,
		Limit:      limit,
		CustomerID: strings.TrimSpace(c.Query("customerID")),
		InvoiceID:  strings.TrimSpace(c.Query("invoiceID")),
	}
	if query.CustomerID == "" {
		query.CustomerID = strings.TrimSpace(c.Query("customerId"))
	}
	if query.InvoiceID == "" {
		query.InvoiceID = strings.TrimSpace(c.Query("invoiceId"))
	}

	fromDateStr := strings.TrimSpace(c.Query("fromDate"))
	toDateStr := strings.TrimSpace(c.Query("toDate"))
	if fromDateStr != "" {
		fromDate, parseErr := time.ParseInLocation("2006-01-02", fromDateStr, location)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "fromDate mora biti u formatu YYYY-MM-DD"})
			return
		}
		query.FromDate = &fromDate
	}
	if toDateStr != "" {
		toDate, parseErr := time.ParseInLocation("2006-01-02", toDateStr, location)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "toDate mora biti u formatu YYYY-MM-DD"})
			return
		}
		end := toDate.AddDate(0, 0, 1)
		query.ToDate = &end
	}
	if query.FromDate != nil && query.ToDate != nil && query.FromDate.After(*query.ToDate) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "fromDate ne može biti poslije toDate"})
		return
	}

	refunds, total, err := repositories.ListRefunds(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju povrata"})
		return
	}

	items := make([]dto.RefundListItemResponse, 0, len(refunds))
	for _, refund := range refunds {
		items = append(items, mapRefundListItem(refund))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.PaginatedRefundsResponse{
		Refunds: items,
		Pagination: dto.RefundPaginationResponse{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}
