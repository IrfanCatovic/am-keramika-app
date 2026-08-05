package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/utils"

	"github.com/gin-gonic/gin"
)

func mapProductResponse(product models.Product) dto.ProductResponse {
	response := dto.ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Slug:          product.Slug,
		Description:   product.Description,
		CategoryID:    product.CategoryID,
		GroupID:       product.GroupID,
		Unit:          product.Unit,
		SalePrice:     product.SalePrice,
		StockQuantity: product.StockQuantity,
		IsActive:      product.IsActive,
	}

	if product.Category.ID != 0 {
		response.Category = &dto.ProductCategorySummaryResponse{
			ID:   product.Category.ID,
			Name: product.Category.Name,
			Slug: product.Category.Slug,
		}
	}

	if product.Group != nil {
		response.Group = &dto.ProductGroupSummaryResponse{
			ID:   product.Group.ID,
			Name: product.Group.Name,
			Slug: product.Group.Slug,
		}
	}

	return response
}

func isProductValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "kategorija nije pronađena") ||
		strings.Contains(msg, "grupa proizvoda nije pronađena") ||
		strings.Contains(msg, "grupa ne pripada izabranoj kategoriji")
}

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	product := models.Product{
		Name:          req.Name,
		Slug:          slug,
		CategoryID:    req.CategoryID,
		GroupID:       req.GroupID,
		Unit:          req.Unit,
		SalePrice:     req.SalePrice,
		StockQuantity: req.StockQuantity,
		Description:   req.Description,
		IsActive:      true,
	}

	err := repositories.CreateProduct(&product)
	if err != nil {
		status := http.StatusInternalServerError
		if isProductValidationError(err) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": "Greska pri kreiranju proizvoda", "error": err.Error()})
		return
	}

	created, err := repositories.GetProductById(strconv.FormatUint(uint64(product.ID), 10))
	if err != nil {
		c.JSON(http.StatusCreated, mapProductResponse(product))
		return
	}

	c.JSON(http.StatusCreated, mapProductResponse(*created))
}

func GetAllProducts(c *gin.Context) {
	search := c.Query("search")
	categoryID := c.Query("categoryID")
	if categoryID == "" {
		categoryID = c.Query("categoryId")
	}

	products, err := repositories.GetAllProducts(search, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Greska pri ucitavanju proizvoda",
			"error":   err.Error(),
		})
		return
	}

	response := make([]dto.ProductResponse, 0, len(products))
	for _, product := range products {
		response = append(response, mapProductResponse(product))
	}

	c.JSON(http.StatusOK, response)
}

func GetProductById(c *gin.Context) {
	id := c.Param("id")

	product, err := repositories.GetProductById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Proizvod nije pronadjen",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, mapProductResponse(*product))
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := repositories.GetProductById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Proizvod nije pronadjen",
			"error":   err.Error(),
		})
		return
	}

	var req dto.UpdateProductRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	product.Name = req.Name
	product.Slug = slug
	product.CategoryID = req.CategoryID
	product.GroupID = req.GroupID
	product.Unit = req.Unit
	product.SalePrice = req.SalePrice
	product.StockQuantity = req.StockQuantity
	product.Description = req.Description

	err = repositories.UpdateProduct(product)
	if err != nil {
		status := http.StatusInternalServerError
		if isProductValidationError(err) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"message": "Greska pri azuriranju proizvoda",
			"error":   err.Error(),
		})
		return
	}

	updated, err := repositories.GetProductById(id)
	if err != nil {
		c.JSON(http.StatusOK, mapProductResponse(*product))
		return
	}

	c.JSON(http.StatusOK, mapProductResponse(*updated))
}

func DeactivateProduct(c *gin.Context) {
	id := c.Param("id")
	err := repositories.DeactivateProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Greska pri deaktiviranju proizvoda",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Proizvod je deaktiviran",
	})
}
