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

func buildPaymentResponse(payment models.Payment) dto.PaymentResponse {
	response := dto.PaymentResponse{
		ID:              payment.ID,
		CustomerID:      payment.CustomerID,
		CreatedByUserID: payment.CreatedByUserID,
		TotalAmount:     payment.TotalAmount,
		CreatedAt:       payment.CreatedAt.Format("2006-01-02 15:04"),
		Allocations:     []dto.PaymentAllocationResponse{},
	}

	if payment.Customer != nil {
		response.Customer = &dto.PaymentCustomerResponse{
			ID:       payment.Customer.ID,
			Name:     payment.Customer.Name,
			Phone:    payment.Customer.Phone,
			IsActive: payment.Customer.IsActive,
		}
	}
	if payment.CreatedByUser.ID != 0 {
		response.CreatedByUser = &dto.PaymentUserResponse{
			ID:       payment.CreatedByUser.ID,
			Username: payment.CreatedByUser.Username,
		}
	}

	for _, allocation := range payment.Allocations {
		response.Allocations = append(response.Allocations, dto.PaymentAllocationResponse{
			ID:        allocation.ID,
			InvoiceID: allocation.InvoiceID,
			Amount:    allocation.Amount,
			Invoice: dto.PaymentAllocationInvoiceResponse{
				ID:          allocation.Invoice.ID,
				TotalAmount: allocation.Invoice.TotalAmount,
				PaidAmount:  allocation.Invoice.PaidAmount,
				Status:      string(allocation.Invoice.Status),
			},
		})
	}
	return response
}

func CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispavan zahtev za uplatu", "error": err.Error()})
		return
	}

	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	payment, err := repositories.CreatePayment(req, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "vec placen") ||
			strings.Contains(errorMessage, "storniran") ||
			strings.Contains(errorMessage, "ne pripada") ||
			strings.Contains(errorMessage, "ne postoji") ||
			strings.Contains(errorMessage, "nije aktivan") ||
			strings.Contains(errorMessage, "iznos") ||
			strings.Contains(errorMessage, "dva puta") ||
			strings.Contains(errorMessage, "bar jedan") ||
			strings.Contains(errorMessage, "ne poklapa") ||
			strings.Contains(errorMessage, "negativan") {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": errorMessage})
		return
	}
	response := buildPaymentResponse(payment)
	c.JSON(http.StatusOK, gin.H{"message": "Uplata je uspesno kreirana", "data": response})
}

func GetAllPayments(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	customerID := c.Query("customerID")
	if customerID == "" {
		customerID = c.Query("customerId")
	}
	fromDateParam := c.Query("fromDate")
	toDateParam := c.Query("toDate")

	page := 1
	limit := 20

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

	payments, total, err := repositories.GetAllPayments(repositories.PaymentListQuery{
		Page:       page,
		Limit:      limit,
		CustomerID: customerID,
		FromDate:   fromDate,
		ToDate:     toDateExclusive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo dobavljanje uplata"})
		return
	}

	response := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		response = append(response, buildPaymentResponse(payment))
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	c.JSON(http.StatusOK, dto.PaginatedPaymentResponse{
		Data:       response,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func GetCustomerPayments(c *gin.Context) {
	customerIDParam := c.Param("id")

	customerIDUint64, err := strconv.ParseUint(customerIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan format kupca ID-ja"})
		return
	}

	customerID := uint(customerIDUint64)

	payments, err := repositories.GetPaymentsByCustomerID(customerID)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if strings.Contains(err.Error(), "kupac nije pronađen") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	response := []dto.PaymentResponse{}
	for _, payment := range payments {
		response = append(response, buildPaymentResponse(payment))
	}
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func GetPaymentByID(c *gin.Context) {
	paymentIDParam := c.Param("id")
	paymentIDUint64, err := strconv.ParseUint(paymentIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan ID uplate"})
		return
	}

	paymentID := uint(paymentIDUint64)
	payment, err := repositories.GetPaymentByID(paymentID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "uplata nije pronađena") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	response := buildPaymentResponse(*payment)
	c.JSON(http.StatusOK, gin.H{"data": response, "message": "Uplata je uspesno ucitana"})
}
