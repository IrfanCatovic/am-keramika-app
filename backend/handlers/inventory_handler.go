package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func mapInventoryMovementResponse(movement models.InventoryMovement) dto.InventoryMovementResponse {
	response := dto.InventoryMovementResponse{
		ID:           movement.ID,
		ProductID:    movement.ProductID,
		ProductName:  movement.Product.Name,
		ProductUnit:  movement.Product.Unit,
		MovementType: movement.MovementType,
		Quantity:     movement.Quantity,
		Note:         movement.Note,
		CreatedAt:    movement.CreatedAt.Format("2006-01-02 15:04"),
	}

	if movement.CreatedByUser.ID != 0 {
		response.CreatedByUser = &dto.InventoryMovementUserResponse{
			ID:       movement.CreatedByUser.ID,
			Username: movement.CreatedByUser.Username,
			FullName: strings.TrimSpace(movement.CreatedByUser.FullName),
		}
	}

	return response
}

func parseInventoryPagination(c *gin.Context, defaultPage, defaultLimit, maxLimit int) (page int, limit int, ok bool) {
	page = defaultPage
	if pageStr := c.Query("page"); pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "page mora biti pozitivan broj"})
			return 0, 0, false
		}
		page = parsed
	}

	limit = defaultLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit mora biti pozitivan broj"})
			return 0, 0, false
		}
		if parsed > maxLimit {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit ne sme biti veći od 100"})
			return 0, 0, false
		}
		limit = parsed
	}

	return page, limit, true
}

func GetLowStock(c *gin.Context) {
	if _, err := auth.GetRole(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	page, limit, ok := parseInventoryPagination(
		c,
		repositories.DefaultLowStockPage,
		repositories.DefaultLowStockLimit,
		repositories.MaxLowStockLimit,
	)
	if !ok {
		return
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
	excludeOutOfStock := c.Query("excludeOutOfStock") == "true"

	products, total, err := repositories.ListLowStockProducts(repositories.LowStockQuery{
		Page:              page,
		Limit:             limit,
		Search:            search,
		CategoryID:        categoryID,
		GroupID:           groupID,
		ExcludeOutOfStock: excludeOutOfStock,
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

func GetInventorySummary(c *gin.Context) {
	if _, err := auth.GetRole(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	lowCount, err := repositories.CountLowStockProducts(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju sažetka lagera"})
		return
	}

	outCount, err := repositories.CountOutOfStockProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju sažetka lagera"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lowStockCount":  lowCount,
		"outOfStockCount": outCount,
	})
}

func GetInventoryMovements(c *gin.Context) {
	if _, err := auth.GetRole(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	page, limit, ok := parseInventoryPagination(
		c,
		repositories.DefaultMovementListPage,
		repositories.DefaultMovementListLimit,
		repositories.MaxMovementListLimit,
	)
	if !ok {
		return
	}

	productID := strings.TrimSpace(c.Query("productID"))
	if productID == "" {
		productID = strings.TrimSpace(c.Query("productId"))
	}
	movementType := strings.TrimSpace(c.Query("type"))
	if movementType == "" {
		movementType = strings.TrimSpace(c.Query("movementType"))
	}
	fromDate := strings.TrimSpace(c.Query("fromDate"))
	toDate := strings.TrimSpace(c.Query("toDate"))

	if fromDate != "" {
		if _, err := time.Parse("2006-01-02", fromDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "fromDate mora biti u formatu YYYY-MM-DD"})
			return
		}
	}
	if toDate != "" {
		if _, err := time.Parse("2006-01-02", toDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "toDate mora biti u formatu YYYY-MM-DD"})
			return
		}
	}

	movements, total, err := repositories.ListInventoryMovements(repositories.MovementListQuery{
		Page:         page,
		Limit:        limit,
		ProductID:    productID,
		MovementType: movementType,
		FromDate:     fromDate,
		ToDate:       toDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju istorije lagera", "error": err.Error()})
		return
	}

	response := make([]dto.InventoryMovementResponse, 0, len(movements))
	for _, movement := range movements {
		response = append(response, mapInventoryMovementResponse(movement))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.PaginatedInventoryMovementsResponse{
		Movements: response,
		Pagination: dto.InventoryMovementPaginationResponse{
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspelo bindovanje JSON-a"})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	err = repositories.AddStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspelo dodavanje stoka", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stok dodan", "data": req})
}

func AdjustStock(c *gin.Context) {
	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan zahtjev", "error": err.Error()})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	if req.NewQuantity == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "newQuantity je obavezno"})
		return
	}

	result, err := repositories.AdjustStock(req.ProductID, *req.NewQuantity, req.Note, createdByUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "proizvod nije pronađen" {
			c.JSON(http.StatusNotFound, gin.H{"message": "Proizvod nije pronađen"})
			return
		}
		if err.Error() == "proizvod nije aktivan" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Proizvod nije aktivan"})
			return
		}
		if err.Error() == "nova količina ne sme biti negativna" {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Korekcija lagera nije uspela", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stanje lagera je ažurirano",
		"data": dto.AdjustStockResponse{
			ProductID:     result.ProductID,
			PreviousStock: result.PreviousStock,
			NewStock:      result.NewStock,
			Change:        result.Change,
			MovementID:    result.MovementID,
		},
	})
}

func SellStock(c *gin.Context) {
	var req dto.SellStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspelo bindovanje JSON-a", "error": err.Error()})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	result, err := repositories.SellStock(req.ProductID, req.Quantity, req.Note, createdByUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspelo prodajanje stoka", "error": err.Error()})
		return
	}

	response := gin.H{
		"message": "Prodaja uspešno evidentirana",
	}

	if result.Warning != "" {
		response["warning"] = result.Warning
	}
	c.JSON(http.StatusOK, response)
}
