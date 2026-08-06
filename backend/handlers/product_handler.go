package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/utils"

	"github.com/gin-gonic/gin"
)

func mapProductResponse(product models.Product, role string) dto.ProductResponse {
	return mapProductListResponse(product, role, nil)
}

func mapProductListResponse(product models.Product, role string, primary *models.ProductImage) dto.ProductResponse {
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
		PrimaryImage:  nil,
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

	if models.CanViewSensitiveProductFields(role) {
		response.PurchasePrice = product.PurchasePrice
		response.MarginPercent = product.MarginPercent
	}

	if primary != nil {
		img := mapProductImageResponse(*primary)
		response.PrimaryImage = &img
	}

	return response
}

func mapProductDetailResponse(product models.Product, role string) dto.ProductResponse {
	response := mapProductListResponse(product, role, nil)
	if len(product.Images) > 0 {
		response.Images = make([]dto.ProductImageResponse, 0, len(product.Images))
		for _, image := range product.Images {
			response.Images = append(response.Images, mapProductImageResponse(image))
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

func rejectWorkerSensitiveProductFields(c *gin.Context, purchasePrice, marginPercent *float64) bool {
	role, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return true
	}
	if role == models.RoleWorker && (purchasePrice != nil || marginPercent != nil) {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Radnik ne smije unositi ili mijenjati nabavnu cijenu ni maržu",
		})
		return true
	}
	return false
}

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	if rejectWorkerSensitiveProductFields(c, req.PurchasePrice, req.MarginPercent) {
		return
	}

	role, _ := auth.GetRole(c)

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

	if models.CanViewSensitiveProductFields(role) {
		product.PurchasePrice = req.PurchasePrice
		product.MarginPercent = req.MarginPercent
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
		c.JSON(http.StatusCreated, mapProductResponse(product, role))
		return
	}

		c.JSON(http.StatusCreated, mapProductDetailResponse(*created, role))
}

func GetAllProducts(c *gin.Context) {
	search := c.Query("search")
	categoryID := c.Query("categoryID")
	if categoryID == "" {
		categoryID = c.Query("categoryId")
	}

	role, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	products, err := repositories.GetAllProducts(search, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Greska pri ucitavanju proizvoda",
			"error":   err.Error(),
		})
		return
	}

	productIDs := make([]uint, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}
	primaryImages, err := repositories.GetPrimaryImagesForProducts(productIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Greska pri ucitavanju slika proizvoda",
			"error":   err.Error(),
		})
		return
	}

	response := make([]dto.ProductResponse, 0, len(products))
	for _, product := range products {
		primary := primaryImages[product.ID]
		var primaryPtr *models.ProductImage
		if primary.ID != 0 {
			img := primary
			primaryPtr = &img
		}
		response = append(response, mapProductListResponse(product, role, primaryPtr))
	}

	c.JSON(http.StatusOK, response)
}

func GetProductById(c *gin.Context) {
	id := c.Param("id")
	role, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	product, err := repositories.GetProductById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Proizvod nije pronadjen",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, mapProductDetailResponse(*product, role))
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

	if rejectWorkerSensitiveProductFields(c, req.PurchasePrice, req.MarginPercent) {
		return
	}

	role, _ := auth.GetRole(c)

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	product.Name = req.Name
	product.Slug = slug
	product.CategoryID = req.CategoryID
	product.Unit = req.Unit
	product.SalePrice = req.SalePrice
	product.StockQuantity = req.StockQuantity
	product.Description = req.Description

	if req.GroupID.Present {
		product.GroupID = req.GroupID.Value
		product.Group = nil
	}

	if models.CanViewSensitiveProductFields(role) {
		product.PurchasePrice = req.PurchasePrice
		product.MarginPercent = req.MarginPercent
	}

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
		c.JSON(http.StatusOK, mapProductResponse(*product, role))
		return
	}

		c.JSON(http.StatusOK, mapProductDetailResponse(*updated, role))
}

func DeactivateProduct(c *gin.Context) {
	id := c.Param("id")
	err := repositories.DeactivateProduct(id)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, repositories.ErrProductHasImages) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"message": "Greska pri deaktiviranju proizvoda",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Proizvod je deaktiviran",
	})
}
