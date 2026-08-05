package handlers

import (
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomerRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer bind data", "error": err.Error()})
		return
	}

	customer := models.Customer{ //pretvaramo request u model jer repository ocekujem model.Customer a ne dto.CreateCustomerRequest
		Name:  req.Name,
		Phone: req.Phone,
	}
	err = repositories.CreateCustomer(&customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create customer", "error": err.Error()})
		return
	}
	response := dto.CustomerResponse{ //pravimo rucno response da bi vratili klijentu sta je neophodno, a ne sve iz modela
		ID:    customer.ID,
		Name:  customer.Name,
		Phone: customer.Phone,
	}
	c.JSON(http.StatusOK, gin.H{"customer": response})
	return
}

func GetAllCustomers(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, limit := 1, 20
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

	customers, total, err := repositories.GetAllCustomers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get customers", "error": err.Error()})
		return
	}

	response := []dto.CustomerListResponse{}
	for _, customer := range customers {
		response = append(response, dto.CustomerListResponse{
			ID:    customer.ID,
			Name:  customer.Name,
			Phone: customer.Phone,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit))) // zaokruzuje broj strana na vecu npr 63 / 20 = 3.15, pa je 4 strana
	c.JSON(http.StatusOK, dto.PaginatedCustomerResponse{
		Data:       response,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
	return
}

func GetCustomerByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer ID", "error": err.Error()})
		return
	}
	customer, err := repositories.GetCustomerByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get customer", "error": err.Error()})
		return
	}
	response := dto.CustomerDetailsResponse{
		ID:       customer.ID,
		Name:     customer.Name,
		Phone:    customer.Phone,
		Debt:     0,
		Invoices: []dto.CustomerInvoiceResponse{},
	}

	for _, invoice := range customer.Invoices {
		response.Invoices = append(response.Invoices, dto.CustomerInvoiceResponse{
			ID:          invoice.ID,
			TotalAmount: invoice.TotalAmount,
			Status:      string(invoice.Status),
			CreatedAt:   invoice.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, response)
	return
}

func GetCustomerFinancialSummary(c *gin.Context) {
	customerIDParam := c.Param("id")
	customerIDUint64, err := strconv.ParseUint(customerIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "neispravan ID kupca"})
		return
	}
	customerID := uint(customerIDUint64)
	summary, err := repositories.GetCustomerFinancialSummary(customerID)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if strings.Contains(err.Error(), "kupac nije pronadjen") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary, "message": "Finansijski pregled kupca je uspesno ucitan"})
	return
}
