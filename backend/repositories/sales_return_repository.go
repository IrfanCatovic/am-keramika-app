package repositories

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateSalesReturnResult struct {
	SalesReturn models.SalesReturn
	RefundID    *uint
}

func CreateSalesReturn(req dto.CreateSalesReturnRequest, createdByUserID uint) (*CreateSalesReturnResult, error) {
	if req.CashRefunded == nil {
		return nil, pricing.ErrSalesReturnCashRefundedRequired
	}
	cashRefunded := *req.CashRefunded

	if len(req.Items) == 0 {
		return nil, pricing.ErrSalesReturnEmptyItems
	}

	seen := make(map[uint]struct{}, len(req.Items))
	productIDs := make([]uint, 0, len(req.Items))
	lines := make([]pricing.SalesReturnLineInput, 0, len(req.Items))
	for _, item := range req.Items {
		if _, exists := seen[item.ProductID]; exists {
			return nil, pricing.ErrSalesReturnDuplicateProduct
		}
		seen[item.ProductID] = struct{}{}
		productIDs = append(productIDs, item.ProductID)
		lines = append(lines, pricing.SalesReturnLineInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	tx := database.DB.Begin()

	lockedProducts := make(map[uint]models.Product, len(productIDs))
	snapshots := make(map[uint]pricing.SalesReturnProductSnapshot, len(productIDs))
	for _, productID := range productIDs {
		var product models.Product
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID).Error
		if err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, pricing.ErrSalesReturnProductNotFound
			}
			return nil, err
		}
		lockedProducts[productID] = product
		snapshots[productID] = pricing.SalesReturnProductSnapshot{
			ID:              product.ID,
			Name:            product.Name,
			Unit:            product.Unit,
			SaleByPackage:   product.SaleByPackage,
			PackageQuantity: product.PackageQuantity,
		}
	}

	prepared, err := pricing.PrepareSalesReturn(strings.TrimSpace(req.Description), lines, snapshots)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	salesReturn := models.SalesReturn{
		TotalAmount:     prepared.TotalAmount,
		Description:     prepared.Description,
		CashRefunded:    cashRefunded,
		CreatedByUserID: createdByUserID,
	}
	if err := tx.Create(&salesReturn).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range prepared.Items {
		row := models.SalesReturnItem{
			SalesReturnID:   salesReturn.ID,
			ProductID:       item.ProductID,
			ProductName:     item.ProductName,
			Unit:            item.Unit,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalPrice:      item.TotalPrice,
			SaleByPackage:   item.SaleByPackage,
			PackageQuantity: item.PackageQuantity,
		}
		if err := tx.Create(&row).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		product := lockedProducts[item.ProductID]
		newStock := pricing.RoundQuantity(product.StockQuantity + item.Quantity)
		if err := tx.Model(&models.Product{}).Where("id = ?", product.ID).
			Update("stock_quantity", newStock).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		product.StockQuantity = newStock
		lockedProducts[item.ProductID] = product

		movement := models.InventoryMovement{
			ProductID:       item.ProductID,
			CreatedByUserID: createdByUserID,
			MovementType:    "return",
			Quantity:        item.Quantity,
			Note:            salesReturnMovementNote(salesReturn.ID, prepared.Description),
		}
		if err := tx.Create(&movement).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	var refundID *uint
	if cashRefunded {
		refund := models.Refund{
			SalesReturnID:   &salesReturn.ID,
			CreatedByUserID: createdByUserID,
			Amount:          prepared.TotalAmount,
			Reason:          salesReturnRefundReason(salesReturn.ID, prepared.Description),
		}
		if err := tx.Create(&refund).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		id := refund.ID
		refundID = &id
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	var created models.SalesReturn
	if err := database.DB.
		Preload("CreatedByUser").
		Preload("Items").
		First(&created, salesReturn.ID).Error; err != nil {
		return nil, err
	}

	return &CreateSalesReturnResult{
		SalesReturn: created,
		RefundID:    refundID,
	}, nil
}

const (
	DefaultSalesReturnListPage  = 1
	DefaultSalesReturnListLimit = 20
	MaxSalesReturnListLimit     = 100
)

var ErrSalesReturnNotFound = errors.New("povrat robe nije pronađen")

type SalesReturnListQuery struct {
	Page         int
	Limit        int
	Search       string
	CashRefunded *bool
}

type SalesReturnListRow struct {
	SalesReturn models.SalesReturn
	ItemsCount  int
}

func ListSalesReturns(q SalesReturnListQuery) ([]SalesReturnListRow, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultSalesReturnListPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultSalesReturnListLimit
	}
	if q.Limit > MaxSalesReturnListLimit {
		q.Limit = MaxSalesReturnListLimit
	}

	var total int64
	if err := buildSalesReturnListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var returns []models.SalesReturn
	offset := (q.Page - 1) * q.Limit
	err := buildSalesReturnListQuery(q).
		Preload("CreatedByUser").
		Order("sales_returns.created_at DESC, sales_returns.id DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&returns).Error
	if err != nil {
		return nil, 0, err
	}

	counts, err := salesReturnItemCounts(salesReturnIDs(returns))
	if err != nil {
		return nil, 0, err
	}

	rows := make([]SalesReturnListRow, 0, len(returns))
	for _, sr := range returns {
		rows = append(rows, SalesReturnListRow{
			SalesReturn: sr,
			ItemsCount:  counts[sr.ID],
		})
	}
	return rows, total, nil
}

func GetSalesReturnByID(id uint) (*CreateSalesReturnResult, error) {
	var salesReturn models.SalesReturn
	err := database.DB.
		Preload("CreatedByUser").
		Preload("Items").
		First(&salesReturn, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesReturnNotFound
		}
		return nil, err
	}

	refundID, err := lookupSalesReturnRefundID(salesReturn.ID)
	if err != nil {
		return nil, err
	}

	return &CreateSalesReturnResult{
		SalesReturn: salesReturn,
		RefundID:    refundID,
	}, nil
}

func buildSalesReturnListQuery(q SalesReturnListQuery) *gorm.DB {
	query := database.DB.Model(&models.SalesReturn{})

	if q.CashRefunded != nil {
		query = query.Where("sales_returns.cash_refunded = ?", *q.CashRefunded)
	}

	search := strings.ToLower(strings.TrimSpace(q.Search))
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where(
			`LOWER(sales_returns.description) LIKE ? OR EXISTS (
				SELECT 1 FROM sales_return_items
				WHERE sales_return_items.sales_return_id = sales_returns.id
					AND sales_return_items.deleted_at IS NULL
					AND LOWER(sales_return_items.product_name) LIKE ?
			)`,
			pattern,
			pattern,
		)
	}

	return query
}

func salesReturnIDs(returns []models.SalesReturn) []uint {
	ids := make([]uint, 0, len(returns))
	for _, sr := range returns {
		ids = append(ids, sr.ID)
	}
	return ids
}

func salesReturnItemCounts(ids []uint) (map[uint]int, error) {
	counts := make(map[uint]int, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}

	type itemCountRow struct {
		SalesReturnID uint
		Count         int
	}
	var rows []itemCountRow
	err := database.DB.Model(&models.SalesReturnItem{}).
		Select("sales_return_id, COUNT(*) AS count").
		Where("sales_return_id IN ?", ids).
		Group("sales_return_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.SalesReturnID] = row.Count
	}
	return counts, nil
}

func lookupSalesReturnRefundID(salesReturnID uint) (*uint, error) {
	var refund models.Refund
	err := database.DB.Select("id").Where("sales_return_id = ?", salesReturnID).First(&refund).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	id := refund.ID
	return &id, nil
}

func salesReturnMovementNote(id uint, description string) string {
	note := fmt.Sprintf("Povrat robe #%d", id)
	if description != "" {
		return note + " — " + description
	}
	return note
}

func salesReturnRefundReason(id uint, description string) string {
	reason := fmt.Sprintf("Povrat robe #%d", id)
	if description != "" {
		return reason + " — " + description
	}
	return reason
}
