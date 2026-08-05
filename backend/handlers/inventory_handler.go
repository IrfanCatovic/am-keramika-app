package handlers

import (
	"net/http"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func AddStock(c *gin.Context) {
	var req dto.AddStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspjelo bindovanje JSON-a"})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	err = repositories.AddStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
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

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	err = repositories.AdjustStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo prilagođavanje stoka", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stok prilagođen", "data": req})
}

func SellStock(c *gin.Context) {
	var req dto.SellStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspjelo bindovanje JSON-a", "error": err.Error()})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	result, err := repositories.SellStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo prodajanje stoka", "error": err.Error()})
		return
	}

	response := gin.H{
		"message": "Prodaja uspjesno evidentirana",
	}

	if result.Warning != "" {
		response["warning"] = result.Warning
	}
	c.JSON(http.StatusOK, response)
}
