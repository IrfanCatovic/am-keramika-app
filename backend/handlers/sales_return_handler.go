package handlers

import (
	"errors"
	"net/http"

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
