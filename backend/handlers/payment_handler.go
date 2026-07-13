package handlers

import (
	"strconv"
	"strings"
	"net/http"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"github.com/gin-gonic/gin"
)

func buildPaymentResponse(payment models.Payment) dto.PaymentResponse {
	response := dto.PaymentResponse{
		ID: payment.ID,
		CustomerID: payment.CustomerID,
		CreatedByUserID: payment.CreatedByUserID,
		TotalAmount: payment.TotalAmount,
		CreatedAt: payment.CreatedAt.Format("2006-01-02 15:04"),
		Allocations: []dto.PaymentAllocationResponse{},
	}

	for _, allocation := range payment.Allocations {
		response.Allocations = append(response.Allocations, dto.PaymentAllocationResponse{
			ID: allocation.ID,
			InvoiceID: allocation.InvoiceID,
			Amount: allocation.Amount,
			Invoice: dto.PaymentAllocationInvoiceResponse{
				ID: allocation.Invoice.ID,
				TotalAmount: allocation.Invoice.TotalAmount,
				PaidAmount: allocation.Invoice.PaidAmount,
				Status: string(allocation.Invoice.Status),
			},
		})
	}
	return response
}

func CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispavan zahtev za uplatu" ,"error": err.Error(),
	})
	return
	}

	// userIDValue, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autorizovan"})
	// 	return
	// }

	// userID, ok := userIDValue.(uint)
	// if !ok {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"message": "Neispravan format korisnickog ID-ja"})
	// 	return
	// }
	userID := uint(7) //createdByUserID privremeni onda ide ovo gore kada napravimo auth middleware

	payment, err := repositories.CreatePayment(req, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
			errorMessage := err.Error()
			if strings.Contains(errorMessage, "vec placen") ||
				strings.Contains(errorMessage, "storniran") ||
				strings.Contains(errorMessage, "ne pripada") ||
				strings.Contains(errorMessage, "ne postoji") ||
				strings.Contains(errorMessage, "iznos") {
				statusCode = http.StatusBadRequest //ovaj ceo error radi samo da bi bratili greske 400 jer to su greske korisnika, a ne 500 neocekivane greske
	}
	c.JSON(statusCode, gin.H{
		"error": errorMessage,
	})
	return
	}
	response := buildPaymentResponse(payment)
	c.JSON(http.StatusOK, gin.H{"message": "Uplata je uspesno kreirana", "data": response})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan ID uplate",})
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