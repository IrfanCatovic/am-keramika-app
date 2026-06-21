package handlers

import(
	"am-keramika-backend/repositories"
	"am-keramika-backend/dto"
	"net/http"
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