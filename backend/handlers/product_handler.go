package handlers

import (
	"net/http"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"github.com/gin-gonic/gin"
	"am-keramika-backend/dto"
)

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	product := models.Product{
		Name: req.Name,
		CategoryID: req.CategoryID,
		Unit: req.Unit,
		SalePrice: req.SalePrice,
		StockQuantity: req.StockQuantity,
		Description: req.Description,
	}

	err := repositories.CreateProduct(&product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greska pri kreiranju proizvoda", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

func GetAllProducts(c *gin.Context) {
	products, err := repositories.GetAllProducts()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Greska pri ucitavanju proizvoda",
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, products)
}

func GetProductById(c *gin.Context) {
	id := c.Param("id")

	product, err := repositories.GetProductById(id)
	if err != nil {
		c.JSON(404, gin.H{
			"message": "Proizvod nije pronadjen",
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, product)
}



func UpdateProduct(c *gin.Context) {

	id := c.Param("id")
	product, err := repositories.GetProductById(id)
	if err != nil {
		c.JSON(404, gin.H{
			"message": "Proizvod nije pronadjen",
			"error": err.Error(),
		})
		return
	}

	var req dto.UpdateProductRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	product.Name = req.Name
	product.CategoryID = req.CategoryID
	product.Unit = req.Unit
	product.SalePrice = req.SalePrice
	product.StockQuantity = req.StockQuantity
	product.Description = req.Description

	err = repositories.UpdateProduct(product)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Greska pri azuriranju proizvoda",
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, product)
}

func DeactivateProduct(c *gin.Context) {
	id := c.Param("id")
	err := repositories.DeactivateProduct(id)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Greska pri deaktiviranju proizvoda",
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Proizvod je deaktiviran",
	})
}

