package handlers

import (
	"am-keramika-backend/dto"
	"am-keramika-backend/repositories"
	"net/http"
	"strconv"
	"math"
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
	if invoice.Customer != nil {
		response.Customer = &dto.CustomerResponse{
			ID: invoice.Customer.ID,
			Name: invoice.Customer.Name,
			Phone: invoice.Customer.Phone,
		}
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



func GetAllInvoices(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

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

	invoices, total, err := repositories.GetAllInvoices(page, limit)
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
			Status:       invoice.Status,
			CreatedAt:    invoice.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	c.JSON(http.StatusOK, dto.PaginatedInvoiceResponse{
		Data: response,
		Page: page,
		Limit: limit,
		Total: total,
		TotalPages: totalPages,
	})

}
