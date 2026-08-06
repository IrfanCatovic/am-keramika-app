package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

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

	c.JSON(http.StatusOK, mapInvoiceResponse(*invoice))
}

func GetAllInvoices(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	search := c.Query("search")
	status := c.Query("status")

	sort := c.Query("sort")
	direction := c.Query("direction")

	page := 1
	limit := 20

	if status != "" && !models.IsValidInvoiceStatus(status) { //imamo u model invoce status koji je ovog tipa sa mogucnostima paid i unpaid
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

	invoices, total, err := repositories.GetAllInvoices(page, limit, search, status, sort, direction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo dobavljanje faktura"})
		return
	}

	response := []dto.InvoiceListResponse{}
	for _, invoice := range invoices {
		customerName := ""
		if invoice.Customer != nil {
			customerName = invoice.Customer.Name
		}
		response = append(response, dto.InvoiceListResponse{
			ID:           invoice.ID,
			CustomerName: customerName,
			TotalAmount:  invoice.TotalAmount,
			PaidAmount:   invoice.PaidAmount,
			Status:       string(invoice.Status),
			CreatedAt:    invoice.CreatedAt.Format("2006-01-02 15:04"),
		})
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
