package handlers

import (
	"am-keramika-backend/repositories"
	"am-keramika-backend/dto"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AddStock(c *gin.Context) {
	var req dto.AddStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspjelo bindovanje JSON-a"})
		return
	}

	//Privremeno dok nema autentifikacije
	//kasnije ce se dobiti iz JWT tokena
	createdByUserID := uint(7) //privremeno dok nema autentifikacije

	err := repositories.AddStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo dodavanje stoka", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stok dodan", "data": req})
}

func AdjustStock(c *gin.Context) {
	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspjelo bindovanje JSON-a", "error": err.Error()})
		return
	}

	//Privremeno dok nema autentifikacije
	//kasnije ce se dobiti iz JWT tokena
	createdByUserID := uint(7) //privremeno dok nema autentifikacije

	err := repositories.AdjustStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo prilagođavanje stoka", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stok prilagođen", "data": req})
}

