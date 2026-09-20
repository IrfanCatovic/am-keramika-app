package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/auth"
	"am-keramika-backend/config"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func CreateSalesReturn(c *gin.Context) {
	var req dto.CreateSalesReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proverite unete podatke"})
		return
	}

	createdByUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Korisnik nije autentifikovan"})
		return
	}

	result, err := repositories.CreateSalesReturn(req, createdByUserID)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, pricing.ErrSalesReturnCashRefundedRequired),
			errors.Is(err, pricing.ErrSalesReturnEmptyItems),
			errors.Is(err, pricing.ErrSalesReturnDuplicateProduct),
			errors.Is(err, pricing.ErrSalesReturnInvalidQuantity),
			errors.Is(err, pricing.ErrSalesReturnInvalidUnitPrice),
			errors.Is(err, pricing.ErrSalesReturnProductNotFound):
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapSalesReturnResponse(result.SalesReturn, result.RefundID))
}

func GetSalesReturns(c *gin.Context) {
	if _, err := auth.GetRole(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Korisnik nije autentifikovan"})
		return
	}

	page := repositories.DefaultSalesReturnListPage
	if pageStr := c.Query("page"); pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page mora biti pozitivan broj"})
			return
		}
		page = parsed
	}

	limit := repositories.DefaultSalesReturnListLimit
	pageSizeStr := strings.TrimSpace(c.Query("pageSize"))
	if pageSizeStr == "" {
		pageSizeStr = strings.TrimSpace(c.Query("limit"))
	}
	if pageSizeStr != "" {
		parsed, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pageSize mora biti pozitivan broj"})
			return
		}
		if parsed > repositories.MaxSalesReturnListLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pageSize ne sme biti veći od 100"})
			return
		}
		limit = parsed
	}

	cashRefunded, err := parseOptionalBoolQuery(c.Query("cashRefunded"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cashRefunded mora biti true ili false"})
		return
	}

	rows, total, err := repositories.ListSalesReturns(repositories.SalesReturnListQuery{
		Page:         page,
		Limit:        limit,
		Search:       strings.TrimSpace(c.Query("search")),
		CashRefunded: cashRefunded,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greška pri učitavanju povrata robe"})
		return
	}

	items := make([]dto.SalesReturnListItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSalesReturnListItem(row))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	c.JSON(http.StatusOK, dto.PaginatedSalesReturnResponse{
		Items:      items,
		Page:       page,
		PageSize:   limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func GetSalesReturnByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nije validan"})
		return
	}

	result, err := repositories.GetSalesReturnByID(uint(id))
	if err != nil {
		if errors.Is(err, repositories.ErrSalesReturnNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greška pri učitavanju povrata robe"})
		return
	}

	c.JSON(http.StatusOK, mapSalesReturnResponse(result.SalesReturn, result.RefundID))
}

func mapSalesReturnListItem(row repositories.SalesReturnListRow) dto.SalesReturnListItemResponse {
	item := dto.SalesReturnListItemResponse{
		ID:           row.SalesReturn.ID,
		Description:  row.SalesReturn.Description,
		TotalAmount:  row.SalesReturn.TotalAmount,
		CashRefunded: row.SalesReturn.CashRefunded,
		ItemsCount:   row.ItemsCount,
		CreatedAt:    config.FormatBusinessDateTime(row.SalesReturn.CreatedAt),
	}
	if row.SalesReturn.CreatedByUser.ID != 0 {
		item.CreatedByUser = mapUserSummary(row.SalesReturn.CreatedByUser)
	}
	return item
}

func parseOptionalBoolQuery(raw string) (*bool, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return nil, nil
	}
	if value == "true" {
		v := true
		return &v, nil
	}
	if value == "false" {
		v := false
		return &v, nil
	}
	return nil, errors.New("invalid bool")
}

func mapSalesReturnResponse(salesReturn models.SalesReturn, refundID *uint) dto.SalesReturnResponse {
	items := make([]dto.SalesReturnItemResponse, 0, len(salesReturn.Items))
	for _, item := range salesReturn.Items {
		items = append(items, dto.SalesReturnItemResponse{
			ID:              item.ID,
			ProductID:       item.ProductID,
			ProductName:     item.ProductName,
			Unit:            item.Unit,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalPrice:      item.TotalPrice,
			SaleByPackage:   item.SaleByPackage,
			PackageQuantity: item.PackageQuantity,
		})
	}

	response := dto.SalesReturnResponse{
		ID:           salesReturn.ID,
		Description:  salesReturn.Description,
		TotalAmount:  salesReturn.TotalAmount,
		CashRefunded: salesReturn.CashRefunded,
		RefundID:     refundID,
		CreatedAt:    config.FormatBusinessDateTime(salesReturn.CreatedAt),
		Items:        items,
	}
	if salesReturn.CreatedByUser.ID != 0 {
		response.CreatedByUser = mapUserSummary(salesReturn.CreatedByUser)
	}
	return response
}
