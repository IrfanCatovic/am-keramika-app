package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"
	"am-keramika-backend/repositories"
	"am-keramika-backend/utils"

	"github.com/gin-gonic/gin"
)

func mapProductResponse(product models.Product, role string) dto.ProductResponse {
	return mapProductListResponse(product, role, nil)
}

func mapProductListResponse(product models.Product, role string, primary *models.ProductImage) dto.ProductResponse {
	response := dto.ProductResponse{
		ID:                 product.ID,
		Name:               product.Name,
		Slug:               product.Slug,
		Description:        product.Description,
		CategoryID:         product.CategoryID,
		GroupID:            product.GroupID,
		Unit:               product.Unit,
		SalePrice:          product.SalePrice,
		EffectiveSalePrice: pricing.GetEffectiveSalePrice(product.SalePrice, product.IsOnSale, product.DiscountPercent),
		StockQuantity:      product.StockQuantity,
		MinStockQuantity:   product.MinStockQuantity,
		IsActive:           product.IsActive,
		IsOnSale:           product.IsOnSale,
		DiscountPercent:    product.DiscountPercent,
		ShowOnHomepage:     product.ShowOnHomepage,
		PricingMode:        pricing.DetectMode(product.PurchasePrice, product.MarginPercent, product.VatPercent),
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

	if models.CanViewSensitiveProductFields(role) {
		response.PurchasePrice = product.PurchasePrice
		response.MarginPercent = product.MarginPercent
		response.VatPercent = product.VatPercent
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
		strings.Contains(msg, "kategorija nije aktivna") ||
		strings.Contains(msg, "grupa proizvoda nije pronađena") ||
		strings.Contains(msg, "grupa ne pripada izabranoj kategoriji") ||
		errors.Is(err, pricing.ErrPurchaseRequired) ||
		errors.Is(err, pricing.ErrManualSaleRequired) ||
		errors.Is(err, pricing.ErrNegativePurchasePrice) ||
		errors.Is(err, pricing.ErrNegativeSalePrice) ||
		errors.Is(err, pricing.ErrNegativeMargin) ||
		errors.Is(err, pricing.ErrNegativeVAT) ||
		errors.Is(err, pricing.ErrNegativeDiscount) ||
		errors.Is(err, pricing.ErrDiscountTooHigh) ||
		errors.Is(err, pricing.ErrSaleRequiresDiscount)
}

func rejectWorkerSensitiveProductFields(c *gin.Context, purchasePrice, marginPercent, vatPercent *float64) bool {
	role, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return true
	}
	if role == models.RoleWorker && (purchasePrice != nil || marginPercent != nil || vatPercent != nil) {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Radnik ne sme unositi ili menjati nabavnu cenu, maržu ni PDV",
		})
		return true
	}
	return false
}

func rejectWorkerDiscountField(c *gin.Context, discountPercent *float64) bool {
	role, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return true
	}
	if role == models.RoleWorker && discountPercent != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Radnik ne sme unositi ili menjati procenat popusta",
		})
		return true
	}
	return false
}

func applyPricingResult(product *models.Product, result pricing.Result) {
	product.SalePrice = result.FinalSalePrice
	product.PurchasePrice = result.PurchasePrice
	product.MarginPercent = result.MarginPercent
	product.VatPercent = result.VatPercent
}

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	if rejectWorkerSensitiveProductFields(c, req.PurchasePrice, req.MarginPercent, req.VatPercent) {
		return
	}
	if rejectWorkerDiscountField(c, req.DiscountPercent) {
		return
	}

	role, _ := auth.GetRole(c)

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	pricingInput := pricing.Input{
		SalePrice: req.SalePrice,
	}
	if models.CanViewSensitiveProductFields(role) {
		pricingInput.PurchasePrice = req.PurchasePrice
		pricingInput.MarginPercent = req.MarginPercent
		pricingInput.VatPercent = req.VatPercent
	}

	priced, err := pricing.Calculate(pricingInput)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error(), "error": err.Error()})
		return
	}

	discountPercent := 0.0
	if models.CanViewSensitiveProductFields(role) && req.DiscountPercent != nil {
		discountPercent = *req.DiscountPercent
	}
	if err := pricing.ValidateSaleDiscount(req.IsOnSale, discountPercent); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error(), "error": err.Error()})
		return
	}

	product := models.Product{
		Name:             req.Name,
		Slug:             slug,
		CategoryID:       req.CategoryID,
		GroupID:          req.GroupID,
		Unit:             req.Unit,
		StockQuantity:    req.StockQuantity,
		MinStockQuantity: req.MinStockQuantity,
		Description:      req.Description,
		IsActive:         true,
		IsOnSale:         req.IsOnSale,
		DiscountPercent:  discountPercent,
		ShowOnHomepage:   req.ShowOnHomepage,
	}
	applyPricingResult(&product, priced)

	err = repositories.CreateProduct(&product)
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
	search := strings.TrimSpace(c.Query("search"))
	categoryID := c.Query("categoryID")
	if categoryID == "" {
		categoryID = c.Query("categoryId")
	}
	groupID := c.Query("groupID")
	if groupID == "" {
		groupID = c.Query("groupId")
	}
	ungrouped := c.Query("ungrouped") == "true"
	includeInactive := c.Query("includeInactive") == "true"
	stockStatus := strings.TrimSpace(c.Query("stockStatus"))
	if stockStatus != "" && stockStatus != "out" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "stockStatus mora biti out ili prazan"})
		return
	}

	if groupID != "" && ungrouped {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "groupID i ungrouped=true ne mogu biti korišteni zajedno",
		})
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

	role, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	products, total, err := repositories.ListProducts(repositories.ProductListQuery{
		Search:          search,
		CategoryID:      categoryID,
		GroupID:         groupID,
		Ungrouped:       ungrouped,
		IncludeInactive: includeInactive,
		StockStatus:     stockStatus,
		Page:            page,
		Limit:           limit,
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

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.PaginatedProductListResponse{
		Products: response,
		Pagination: dto.ProductPaginationResponse{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
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

	if rejectWorkerSensitiveProductFields(c, req.PurchasePrice, req.MarginPercent, req.VatPercent) {
		return
	}
	if rejectWorkerDiscountField(c, req.DiscountPercent) {
		return
	}

	role, _ := auth.GetRole(c)

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	existingMode := pricing.DetectMode(product.PurchasePrice, product.MarginPercent, product.VatPercent)

	purchase := product.PurchasePrice
	margin := product.MarginPercent
	vat := product.VatPercent
	sale := &product.SalePrice

	if models.CanViewSensitiveProductFields(role) {
		purchase = req.PurchasePrice
		margin = req.MarginPercent
		vat = req.VatPercent
		sale = req.SalePrice
	} else {
		// Radnik: zadrži skrivene pricing vrijednosti.
		if existingMode == pricing.ModeCalculated {
			if req.SalePrice != nil && *req.SalePrice != product.SalePrice {
				c.JSON(http.StatusForbidden, gin.H{
					"message": "Cena se automatski obračunava; radnik ne sme menjati calculated prodajnu cenu",
					"error":   "Cena se automatski obračunava; radnik ne sme menjati calculated prodajnu cenu",
				})
				return
			}
			sale = &product.SalePrice
		} else if req.SalePrice != nil {
			sale = req.SalePrice
		}
	}

	priced, err := pricing.Calculate(pricing.Input{
		PurchasePrice: purchase,
		MarginPercent: margin,
		VatPercent:    vat,
		SalePrice:     sale,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error(), "error": err.Error()})
		return
	}

	product.Name = req.Name
	product.Slug = slug
	product.CategoryID = req.CategoryID
	product.Unit = req.Unit
	product.StockQuantity = req.StockQuantity
	product.MinStockQuantity = req.MinStockQuantity
	product.Description = req.Description
	applyPricingResult(product, priced)

	if req.GroupID.Present {
		product.GroupID = req.GroupID.Value
		product.Group = nil
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	if req.IsOnSale != nil {
		product.IsOnSale = *req.IsOnSale
	}
	if models.CanViewSensitiveProductFields(role) && req.DiscountPercent != nil {
		product.DiscountPercent = *req.DiscountPercent
	}
	if req.ShowOnHomepage != nil {
		product.ShowOnHomepage = *req.ShowOnHomepage
	}

	if err := pricing.ValidateSaleDiscount(product.IsOnSale, product.DiscountPercent); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error(), "error": err.Error()})
		return
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

func ActivateProduct(c *gin.Context) {
	id := c.Param("id")
	err := repositories.ActivateProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Greska pri aktiviranju proizvoda",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Proizvod je aktiviran"})
}
