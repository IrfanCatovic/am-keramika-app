package handlers

import(
	"am-keramika-backend/repositories"
	"am-keramika-backend/dto"
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"
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

	c.JSON(http.StatusOK, gin.H{"invoice": invoice})

}