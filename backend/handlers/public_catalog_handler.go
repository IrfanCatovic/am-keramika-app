package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func mapPublicProductResponse(product models.Product, primary *models.ProductImage, includeImages bool) dto.PublicProductResponse {
	response := dto.PublicProductResponse{
		ID:                 product.ID,
		Name:               product.Name,
		Slug:               product.Slug,
		Description:        product.Description,
		Unit:               product.Unit,
		SalePrice:          product.SalePrice,
		EffectiveSalePrice: pricing.GetEffectiveSalePrice(product.SalePrice, product.IsOnSale, product.DiscountPercent),
		IsOnSale:           product.IsOnSale,
		DiscountPercent:    product.DiscountPercent,
		InStock:            product.StockQuantity > 0,
		ShowOnHomepage:     product.ShowOnHomepage,
		PrimaryImage:       nil,
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
	if primary != nil {
		img := mapProductImageResponse(*primary)
		response.PrimaryImage = &img
	}
	if includeImages && len(product.Images) > 0 {
		response.Images = make([]dto.ProductImageResponse, 0, len(product.Images))
		for _, image := range product.Images {
			response.Images = append(response.Images, mapProductImageResponse(image))
		}
	}
	return response
}

func GetPublicProducts(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	categoryID := c.Query("categoryID")
	if categoryID == "" {
		categoryID = c.Query("categoryId")
	}
	categorySlug := strings.TrimSpace(c.Query("categorySlug"))
	if categorySlug == "" {
		categorySlug = strings.TrimSpace(c.Query("category"))
	}
	groupID := c.Query("groupID")
	if groupID == "" {
		groupID = c.Query("groupId")
	}
	groupSlug := strings.TrimSpace(c.Query("groupSlug"))
	if groupSlug == "" {
		groupSlug = strings.TrimSpace(c.Query("group"))
	}
	ungrouped := c.Query("ungrouped") == "true"
	onSaleOnly := c.Query("onSale") == "true"
	homepageOnly := c.Query("homepage") == "true" || c.Query("featured") == "true"
	inStockOnly := c.Query("inStock") == "true"
	random := c.Query("random") == "true"
	sortRaw := c.Query("sort")

	if categoryID != "" && categorySlug != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "categoryID i categorySlug ne mogu biti korišteni zajedno"})
		return
	}
	if groupID != "" && groupSlug != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "groupID i groupSlug ne mogu biti korišteni zajedno"})
		return
	}
	if (groupID != "" || groupSlug != "") && ungrouped {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "group filter i ungrouped=true ne mogu biti korišteni zajedno",
		})
		return
	}

	sort := repositories.NormalizePublicSort(sortRaw)
	if sortRaw != "" && sort == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "nepodržan sort parametar"})
		return
	}

	page := repositories.DefaultProductListPage
	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "page mora biti pozitivan broj"})
			return
		}
		page = parsedPage
	}

	limit := repositories.DefaultProductListLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit mora biti pozitivan broj"})
			return
		}
		if parsedLimit > repositories.MaxProductListLimit {
			c.JSON(http.StatusBadRequest, gin.H{"message": "limit ne sme biti veći od 100"})
			return
		}
		limit = parsedLimit
	}

	var excludeID uint
	if excludeStr := strings.TrimSpace(c.Query("excludeId")); excludeStr != "" {
		parsed, err := strconv.ParseUint(excludeStr, 10, 64)
		if err != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "excludeId mora biti pozitivan broj"})
			return
		}
		excludeID = uint(parsed)
	}

	products, total, err := repositories.ListPublicProducts(repositories.PublicProductListQuery{
		Search:       search,
		CategoryID:   categoryID,
		CategorySlug: categorySlug,
		GroupID:      groupID,
		GroupSlug:    groupSlug,
		Ungrouped:    ungrouped,
		OnSaleOnly:   onSaleOnly,
		HomepageOnly: homepageOnly,
		InStockOnly:  inStockOnly,
		ExcludeID:    excludeID,
		Random:       random,
		Sort:         sort,
		Page:         page,
		Limit:        limit,
	})
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

	response := make([]dto.PublicProductResponse, 0, len(products))
	for _, product := range products {
		primary := primaryImages[product.ID]
		var primaryPtr *models.ProductImage
		if primary.ID != 0 {
			img := primary
			primaryPtr = &img
		}
		response = append(response, mapPublicProductResponse(product, primaryPtr, false))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.PaginatedPublicProductListResponse{
		Products: response,
		Pagination: dto.ProductPaginationResponse{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

func GetPublicProductBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug je obavezan"})
		return
	}

	product, err := repositories.GetPublicProductBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Proizvod nije pronadjen",
		})
		return
	}

	var primaryPtr *models.ProductImage
	for i := range product.Images {
		if product.Images[i].IsPrimary {
			img := product.Images[i]
			primaryPtr = &img
			break
		}
	}
	if primaryPtr == nil && len(product.Images) > 0 {
		img := product.Images[0]
		primaryPtr = &img
	}

	c.JSON(http.StatusOK, mapPublicProductResponse(*product, primaryPtr, true))
}

func GetPublicCategories(c *gin.Context) {
	categories, err := repositories.GetCategories(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Greska pri ucitavanju kategorija",
			"error":   err.Error(),
		})
		return
	}

	response := make([]dto.PublicCategoryResponse, 0, len(categories))
	for _, category := range categories {
		if !category.IsActive {
			continue
		}
		response = append(response, dto.PublicCategoryResponse{
			ID:   category.ID,
			Name: category.Name,
			Slug: category.Slug,
		})
	}
	c.JSON(http.StatusOK, response)
}

func GetPublicCategoryBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug je obavezan"})
		return
	}

	category, err := repositories.GetPublicCategoryBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Kategorija nije pronađena"})
		return
	}

	c.JSON(http.StatusOK, dto.PublicCategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	})
}

func GetPublicProductGroups(c *gin.Context) {
	categoryID := c.Query("categoryID")
	if categoryID == "" {
		categoryID = c.Query("categoryId")
	}
	categorySlug := strings.TrimSpace(c.Query("categorySlug"))
	if categorySlug == "" {
		categorySlug = strings.TrimSpace(c.Query("category"))
	}

	if categoryID != "" && categorySlug != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "categoryID i categorySlug ne mogu biti korišteni zajedno"})
		return
	}

	if categorySlug != "" {
		category, err := repositories.GetPublicCategoryBySlug(categorySlug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Kategorija nije pronađena"})
			return
		}
		categoryID = strconv.FormatUint(uint64(category.ID), 10)
	}

	groups, err := repositories.GetAllProductGroups(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Greska pri ucitavanju grupa",
			"error":   err.Error(),
		})
		return
	}

	response := make([]dto.PublicProductGroupResponse, 0, len(groups))
	for _, group := range groups {
		response = append(response, dto.PublicProductGroupResponse{
			ID:         group.ID,
			Name:       group.Name,
			Slug:       group.Slug,
			CategoryID: group.CategoryID,
		})
	}
	c.JSON(http.StatusOK, response)
}
