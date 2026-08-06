package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func missingQuantity(stockQuantity, minStockQuantity float64) float64 {
	missing := minStockQuantity - stockQuantity
	if missing < 0 {
		return 0
	}
	return missing
}

func mapLowStockProductResponse(product models.Product, primary *models.ProductImage) dto.LowStockProductResponse {
	response := dto.LowStockProductResponse{
		ID:               product.ID,
		Name:             product.Name,
		Unit:             product.Unit,
		StockQuantity:    product.StockQuantity,
		MinStockQuantity: product.MinStockQuantity,
		MissingQuantity:  missingQuantity(product.StockQuantity, product.MinStockQuantity),
		Category:         nil,
		Group:            nil,
		PrimaryImage:     nil,
	}

	if product.Category.ID != 0 {
		response.Category = &dto.LowStockCategoryResponse{
			ID:   product.Category.ID,
			Name: product.Category.Name,
		}
	}

	if product.Group != nil {
		response.Group = &dto.LowStockGroupResponse{
			ID:   product.Group.ID,
			Name: product.Group.Name,
		}
	}

	if primary != nil {
		img := mapProductImageResponse(*primary)
		response.PrimaryImage = &img
	}

	return response
}

func GetLowStock(c *gin.Context) {
	if _, err := auth.GetRole(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	page := repositories.DefaultLowStockPage
	if pageStr := c.Query("page"); pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "page mora biti pozitivan broj"})
			return
		}
		page = parsed
	}

	limit := repositories.DefaultLowStockLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit mora biti pozitivan broj"})
			return
		}
		if parsed > repositories.MaxLowStockLimit {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit ne smije biti veći od 100"})
			return
		}
		limit = parsed
	}

	search := strings.TrimSpace(c.Query("search"))
	categoryID := c.Query("categoryID")
	if categoryID == "" {
		categoryID = c.Query("categoryId")
	}
	groupID := c.Query("groupID")
	if groupID == "" {
		groupID = c.Query("groupId")
	}

	products, total, err := repositories.ListLowStockProducts(repositories.LowStockQuery{
		Page:       page,
		Limit:      limit,
		Search:     search,
		CategoryID: categoryID,
		GroupID:    groupID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju low-stock proizvoda", "error": err.Error()})
		return
	}

	productIDs := make([]uint, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}
	primaryImages, err := repositories.GetPrimaryImagesForProducts(productIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju slika proizvoda", "error": err.Error()})
		return
	}

	response := make([]dto.LowStockProductResponse, 0, len(products))
	for _, product := range products {
		primary := primaryImages[product.ID]
		var primaryPtr *models.ProductImage
		if primary.ID != 0 {
			img := primary
			primaryPtr = &img
		}
		response = append(response, mapLowStockProductResponse(product, primaryPtr))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.PaginatedLowStockResponse{
		Products: response,
		Pagination: dto.LowStockPaginationResponse{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

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
