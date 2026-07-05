package handlers

import (

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
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspesno kreiranje uplate" ,"error": err.Error()})
		return
	}

	response := buildPaymentResponse(payment)

	c.JSON(http.StatusOK, gin.H{"message": "Uplata je uspesno kreirana", "data": response})

	
}