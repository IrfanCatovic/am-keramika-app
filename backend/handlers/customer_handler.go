package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer bind data", "error": err.Error()})
		return
	}

	customer := models.Customer{
		Name:  strings.TrimSpace(req.Name),
		Phone: strings.TrimSpace(req.Phone),
	}
	if err := repositories.CreateCustomer(&customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create customer", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer": dto.CustomerResponse{
		ID:    customer.ID,
		Name:  customer.Name,
		Phone: customer.Phone,
	}})
}

func GetAllCustomers(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	search := strings.TrimSpace(c.Query("search"))
	includeInactive := c.Query("includeInactive") == "true"

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

	customers, total, err := repositories.GetAllCustomers(repositories.CustomerListQuery{
		Page:            page,
		Limit:           limit,
		Search:          search,
		IncludeInactive: includeInactive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get customers", "error": err.Error()})
		return
	}

	response := make([]dto.CustomerListResponse, 0, len(customers))
	for _, customer := range customers {
		response = append(response, dto.CustomerListResponse{
			ID:    customer.ID,
			Name:  customer.Name,
			Phone: customer.Phone,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	c.JSON(http.StatusOK, dto.PaginatedCustomerResponse{
		Data:       response,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func GetCustomerByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer ID", "error": err.Error()})
		return
	}

	customer, err := repositories.GetCustomerByID(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repositories.ErrCustomerNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Failed to get customer", "error": err.Error()})
		return
	}

	response := dto.CustomerDetailsResponse{
		ID:       customer.ID,
		Name:     customer.Name,
		Phone:    customer.Phone,
		Debt:     customer.TotalDebt,
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
}

func UpdateCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer ID"})
		return
	}

	customer, err := repositories.GetCustomerByID(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repositories.ErrCustomerNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Failed to get customer", "error": err.Error()})
		return
	}

	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer data", "error": err.Error()})
		return
	}

	customer.Name = strings.TrimSpace(req.Name)
	customer.Phone = strings.TrimSpace(req.Phone)

	if err := repositories.UpdateCustomer(customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update customer", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Customer updated",
		"customer": dto.CustomerResponse{ID: customer.ID, Name: customer.Name, Phone: customer.Phone},
	})
}

func UpdateCustomerStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer ID"})
		return
	}

	var req dto.UpdateCustomerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status data", "error": err.Error()})
		return
	}

	if err := repositories.UpdateCustomerStatus(uint(id), req.IsActive); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, repositories.ErrCustomerNotFound):
			status = http.StatusNotFound
		case errors.Is(err, repositories.ErrCustomerHasOpenInvoices):
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"message": "Failed to update customer status", "error": err.Error()})
		return
	}

	customer, err := repositories.GetCustomerByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get customer", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Customer status updated",
		"customer": dto.CustomerResponse{
			ID:    customer.ID,
			Name:  customer.Name,
			Phone: customer.Phone,
		},
	})
}

func DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer ID"})
		return
	}

	if err := repositories.DeleteCustomer(uint(id)); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, repositories.ErrCustomerNotFound):
			status = http.StatusNotFound
		case errors.Is(err, repositories.ErrCustomerHasHistory):
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"message": "Failed to delete customer", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
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

		if strings.Contains(err.Error(), "kupac nije pronadjen") || strings.Contains(err.Error(), "customer not found") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary, "message": "Finansijski pregled kupca je uspesno ucitan"})
}
