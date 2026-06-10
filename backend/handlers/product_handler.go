package handlers

import (
	"net/http"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	err := repositories.CreateProduct(&product)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Greska pri kreiranju proizvoda",
			"error": err.Error(),
		})
		return
	}

	c.JSON(201, product)
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

	err = c.ShouldBindJSON(product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

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