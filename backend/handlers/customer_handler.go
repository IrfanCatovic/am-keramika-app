package handlers

import (
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"net/http"
	"github.com/gin-gonic/gin"
)

func CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomerRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid customer bind data", "error": err.Error()})
		return
	}

	customer := models.Customer{ //pretvaramo request u model jer repository ocekujem model.Customer a ne dto.CreateCustomerRequest
		Name: req.Name,
		Phone: req.Phone,
	}
	err = repositories.CreateCustomer(&customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create customer", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer": customer})
	return
}

