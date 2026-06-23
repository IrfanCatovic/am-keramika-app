package handlers

import (
	"am-keramika-backend/dto"
	"am-keramika-backend/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateInvoice(c *gin.Context) {
	var req dto.CreateInvoiceRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdByUserID := uint(7)

	invoice, err := repositories.CreateInvoice(req, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoice": invoice})
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

	response := dto.InvoiceResponse{
		ID:          invoice.ID,
		TotalAmount: invoice.TotalAmount,
		Status:      invoice.Status,
		Items:       []dto.InvoiceItemResponse{},
	}

	for _, item := range invoice.Items {
		response.Items = append(response.Items, dto.InvoiceItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.Product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		})
	}

	c.JSON(http.StatusOK, response)
}
