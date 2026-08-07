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

func CreateInvoice(c *gin.Context) {
	var req dto.CreateInvoiceRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Korisnik nije autentifikovan"})
		return
	}

	invoice, err := repositories.CreateInvoice(req, createdByUserID)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		switch {
		case strings.Contains(msg, "kupac nije aktivan"):
			status = http.StatusConflict
		case strings.Contains(msg, "kupac nije pronađen"),
			strings.Contains(msg, "proizvod nije pronađen"),
			strings.Contains(msg, "nema dovoljno"):
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	fullInvoice, err := repositories.GetInvoiceByID(invoice.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Račun nije pronađen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoice": mapInvoiceResponse(*fullInvoice)})
}

func GetInvoiceByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nije validan"})
		return
	}

	invoice, err := repositories.GetInvoiceByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Račun nije pronađen"})
		return
	}

	response := mapInvoiceResponse(*invoice)
	enrichInvoiceCancellationData(&response, invoice.ID, string(invoice.Status))
	c.JSON(http.StatusOK, response)
}

func GetAllInvoices(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	search := strings.TrimSpace(c.Query("search"))
	status := c.Query("status")
	customerID := c.Query("customerID")
	if customerID == "" {
		customerID = c.Query("customerId")
	}
	fromDateParam := c.Query("fromDate")
	toDateParam := c.Query("toDate")
	sort := c.Query("sort")
	direction := c.Query("direction")

	page := 1
	limit := 20

	if status != "" && !models.IsValidInvoiceStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan status fakture"})
		return
	}

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}
	if sort != "" && !models.IsValidInvoiceSort(sort) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan sort fakture"})
		return
	}
	if direction != "" && !models.IsValidSortDirection(direction) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan direction fakture"})
		return
	}

	location, err := time.LoadLocation("Europe/Belgrade")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo učitavanje vremenske zone"})
		return
	}

	var fromDate *time.Time
	var toDateExclusive *time.Time
	var fromDay, toDay time.Time

	if fromDateParam != "" {
		parsed, err := time.ParseInLocation("2006-01-02", fromDateParam, location)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan format fromDate (YYYY-MM-DD)"})
			return
		}
		fromDay = parsed
		fromDate = &parsed
	}
	if toDateParam != "" {
		parsed, err := time.ParseInLocation("2006-01-02", toDateParam, location)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan format toDate (YYYY-MM-DD)"})
			return
		}
		toDay = parsed
		end := parsed.AddDate(0, 0, 1)
		toDateExclusive = &end
	}
	if fromDateParam != "" && toDateParam != "" && toDay.Before(fromDay) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "toDate ne može biti prije fromDate"})
		return
	}

	invoices, total, err := repositories.GetAllInvoices(repositories.InvoiceListQuery{
		Page:       page,
		Limit:      limit,
		Search:     search,
		Status:     status,
		CustomerID: customerID,
		FromDate:   fromDate,
		ToDate:     toDateExclusive,
		Sort:       sort,
		Direction:  direction,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo dobavljanje faktura"})
		return
	}

	response := make([]dto.InvoiceListResponse, 0, len(invoices))
	for _, invoice := range invoices {
		response = append(response, mapInvoiceListResponse(invoice))
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	c.JSON(http.StatusOK, dto.PaginatedInvoiceResponse{
		Data:       response,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func GetCustomerOpenInvoices(c *gin.Context) {
	customerIDParam := c.Param("id")

	customerIDUint64, err := strconv.ParseUint(customerIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan ID kupca"})
		return
	}

	customerID := uint(customerIDUint64)
	invoices, err := repositories.GetOpenInvoicesByCustomerID(customerID)

	if err != nil {
		statusCode := http.StatusInternalServerError

		if strings.Contains(err.Error(), "kupac nije pronađen") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := []dto.InvoiceListResponse{}
	for _, invoice := range invoices {
		remainingAmount := invoice.TotalAmount - invoice.PaidAmount
		response = append(response, dto.InvoiceListResponse{
			ID:              invoice.ID,
			TotalAmount:     invoice.TotalAmount,
			PaidAmount:      invoice.PaidAmount,
			RemainingAmount: remainingAmount,
			Status:          string(invoice.Status),
			CreatedAt:       invoice.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "Otvoreni racuni kupca su uspesno ucitani",
	})
}

func CancelInvoice(c *gin.Context) {
	invoiceIDParam := c.Param("id")
	invoiceIDUint64, err := strconv.ParseUint(invoiceIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan ID fakture"})
		return
	}
	invoiceID := uint(invoiceIDUint64)

	var req dto.CancelInvoiceRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Korisnik nije autentifikovan"})
		return
	}

	invoiceCancellation, err := repositories.CancelInvoice(invoiceID, req, createdByUserID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "vec otkazan") {
			statusCode = http.StatusBadRequest
		}
		if strings.Contains(err.Error(), "record not found") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    invoiceCancellation,
		"message": "Racun je uspesno storniran",
	})
}
