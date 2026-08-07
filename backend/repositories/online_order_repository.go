package repositories

import (
	"errors"
	"math"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/pricing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrOnlineOrderValidation         = errors.New("proverite unesene podatke")
	ErrOnlineOrderHoneypot           = errors.New("zahtev odbijen")
	ErrOnlineOrderInsufficientStock  = errors.New("nema dovoljno proizvoda na stanju za izabranu količinu")
	ErrOnlineOrderProductUnavailable = errors.New("jedan od proizvoda više nije dostupan")
	ErrOnlineOrderTooManyItems       = errors.New("previše stavki u narudžbini")
	ErrOnlineOrderInvalidQuantity    = errors.New("količina mora biti veća od 0")
	ErrOnlineOrderInvalidEmail       = errors.New("email adresa nije ispravna")
	ErrOnlineOrderNotFound           = errors.New("narudžbina nije pronađena")
	ErrOnlineOrderAlreadyProcessed   = errors.New("narudžbina je već obrađena")
	ErrOnlineOrderConfirmStock       = errors.New("nema dovoljno proizvoda na stanju za potvrdu ove narudžbine")
	ErrOnlineOrderConfirmUnavailable = errors.New("jedan od proizvoda više nije dostupan")
	ErrOnlineOrderDeleteNotPending   = errors.New("potvrđena narudžbina se ne može obrisati")
	ErrOnlineOrderConfirmCustomer    = errors.New("izaberite postojećeg kupca ili kreirajte novog")
)

const (
	maxOnlineOrderItems     = 100
	maxOnlineOrderQuantity  = 100_000.0
	maxOnlineOrderFirstName = 100
	maxOnlineOrderLastName  = 100
	maxOnlineOrderPhone     = 50
	maxOnlineOrderCity      = 150
	maxOnlineOrderAddress   = 250
	maxOnlineOrderEmail     = 254
	maxOnlineOrderNote      = 1000
	minOnlineOrderPhoneLen  = 6
)

type OnlineOrderCreateError struct {
	Err       error
	ProductID *uint
	Code      string
}

func (e *OnlineOrderCreateError) Error() string {
	if e.Err == nil {
		return "online order error"
	}
	return e.Err.Error()
}

func (e *OnlineOrderCreateError) Unwrap() error {
	return e.Err
}

func orderCreateErr(err error, productID *uint, code string) *OnlineOrderCreateError {
	return &OnlineOrderCreateError{Err: err, ProductID: productID, Code: code}
}

// CreateOnlineOrder validates products, snapshots effective prices, and persists
// a pending OnlineOrder. It never mutates stock or creates invoices/payments.
func CreateOnlineOrder(req dto.PublicCreateOnlineOrderRequest) (*models.OnlineOrder, error) {
	if strings.TrimSpace(req.Website) != "" {
		return nil, orderCreateErr(ErrOnlineOrderHoneypot, nil, "rejected")
	}

	firstName := clip(strings.TrimSpace(req.FirstName), maxOnlineOrderFirstName)
	lastName := clip(strings.TrimSpace(req.LastName), maxOnlineOrderLastName)
	phone := clip(strings.TrimSpace(req.Phone), maxOnlineOrderPhone)
	city := clip(strings.TrimSpace(req.City), maxOnlineOrderCity)
	address := clip(strings.TrimSpace(req.Address), maxOnlineOrderAddress)
	email := clip(strings.TrimSpace(req.Email), maxOnlineOrderEmail)
	note := clip(strings.TrimSpace(req.Note), maxOnlineOrderNote)

	if firstName == "" || lastName == "" || city == "" || address == "" {
		return nil, orderCreateErr(ErrOnlineOrderValidation, nil, "validation")
	}
	if countPhoneDigits(phone) < minOnlineOrderPhoneLen {
		return nil, orderCreateErr(ErrOnlineOrderValidation, nil, "validation")
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return nil, orderCreateErr(ErrOnlineOrderInvalidEmail, nil, "invalid_email")
		}
	}
	if len(req.Items) == 0 {
		return nil, orderCreateErr(ErrOnlineOrderValidation, nil, "validation")
	}
	if len(req.Items) > maxOnlineOrderItems {
		return nil, orderCreateErr(ErrOnlineOrderTooManyItems, nil, "too_many_items")
	}

	type preparedItem struct {
		productID   uint
		quantity    float64
		productName string
		productSlug string
		unit        string
		unitPrice   float64
		totalPrice  float64
	}

	prepared := make([]preparedItem, 0, len(req.Items))
	seen := make(map[uint]float64, len(req.Items))

	for _, item := range req.Items {
		if item.ProductID == 0 {
			return nil, orderCreateErr(ErrOnlineOrderValidation, nil, "validation")
		}
		if item.Quantity <= 0 || math.IsNaN(item.Quantity) || math.IsInf(item.Quantity, 0) {
			pid := item.ProductID
			return nil, orderCreateErr(ErrOnlineOrderInvalidQuantity, &pid, "invalid_quantity")
		}
		qty := pricing.RoundToTwoDecimals(item.Quantity)
		if qty <= 0 || qty > maxOnlineOrderQuantity {
			pid := item.ProductID
			return nil, orderCreateErr(ErrOnlineOrderInvalidQuantity, &pid, "invalid_quantity")
		}
		seen[item.ProductID] = pricing.RoundToTwoDecimals(seen[item.ProductID] + qty)
	}

	if len(seen) > maxOnlineOrderItems {
		return nil, orderCreateErr(ErrOnlineOrderTooManyItems, nil, "too_many_items")
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var totalAmount float64

	for productID, quantity := range seen {
		var product models.Product
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Category").
			First(&product, productID).Error
		if err != nil {
			tx.Rollback()
			pid := productID
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, orderCreateErr(ErrOnlineOrderProductUnavailable, &pid, "unavailable")
			}
			return nil, err
		}

		if !product.IsActive || product.Category.ID == 0 || !product.Category.IsActive {
			tx.Rollback()
			pid := productID
			return nil, orderCreateErr(ErrOnlineOrderProductUnavailable, &pid, "unavailable")
		}
		if product.StockQuantity < quantity {
			tx.Rollback()
			pid := productID
			return nil, orderCreateErr(ErrOnlineOrderInsufficientStock, &pid, "insufficient_stock")
		}

		unitPrice := pricing.GetEffectiveSalePrice(product.SalePrice, product.IsOnSale, product.DiscountPercent)
		lineTotal := pricing.RoundToTwoDecimals(unitPrice * quantity)
		totalAmount = pricing.RoundToTwoDecimals(totalAmount + lineTotal)

		prepared = append(prepared, preparedItem{
			productID:   product.ID,
			quantity:    quantity,
			productName: product.Name,
			productSlug: product.Slug,
			unit:        product.Unit,
			unitPrice:   unitPrice,
			totalPrice:  lineTotal,
		})
	}

	order := models.OnlineOrder{
		Status:      models.OnlineOrderStatusPending,
		FirstName:   firstName,
		LastName:    lastName,
		Phone:       phone,
		City:        city,
		Address:     address,
		Email:       email,
		Note:        note,
		TotalAmount: totalAmount,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range prepared {
		row := models.OnlineOrderItem{
			OnlineOrderID: order.ID,
			ProductID:     item.productID,
			ProductName:   item.productName,
			ProductSlug:   item.productSlug,
			Unit:          item.unit,
			Quantity:      item.quantity,
			UnitPrice:     item.unitPrice,
			TotalPrice:    item.totalPrice,
		}
		if err := tx.Create(&row).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := database.DB.Preload("Items").First(&order, order.ID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func clip(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func countPhoneDigits(phone string) int {
	n := 0
	for _, r := range phone {
		if unicode.IsDigit(r) {
			n++
		}
	}
	return n
}

func CountPendingOnlineOrders() (int64, error) {
	var count int64
	err := database.DB.Model(&models.OnlineOrder{}).
		Where("status = ?", models.OnlineOrderStatusPending).
		Count(&count).Error
	return count, err
}

type OnlineOrderListQuery struct {
	Page     int
	Limit    int
	Status   string
	Search   string
	FromDate *time.Time
	ToDate   *time.Time
}

func ListOnlineOrders(q OnlineOrderListQuery) ([]models.OnlineOrder, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	query := database.DB.Model(&models.OnlineOrder{})
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}
	if q.Search != "" {
		s := "%" + strings.ToLower(strings.TrimSpace(q.Search)) + "%"
		query = query.Where(
			"(LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(phone) LIKE ?)",
			s, s, s,
		)
	}
	if q.FromDate != nil {
		query = query.Where("created_at >= ?", *q.FromDate)
	}
	if q.ToDate != nil {
		query = query.Where("created_at < ?", *q.ToDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []models.OnlineOrder
	err := query.
		Preload("Items").
		Order("created_at DESC, id DESC").
		Offset((q.Page - 1) * q.Limit).
		Limit(q.Limit).
		Find(&orders).Error
	return orders, total, err
}

func GetOnlineOrderByID(id uint) (*models.OnlineOrder, error) {
	var order models.OnlineOrder
	err := database.DB.Preload("Items").First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOnlineOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

// ConfirmOnlineOrder atomically converts a pending OnlineOrder into an unpaid customer Invoice.
// Invoice item prices come from OnlineOrderItem snapshots — not current product prices.
func ConfirmOnlineOrder(orderID uint, req dto.ConfirmOnlineOrderRequest, confirmedByUserID uint) (*models.OnlineOrder, *models.Invoice, error) {
	hasExisting := req.CustomerID != nil && *req.CustomerID > 0
	hasNew := req.NewCustomer != nil && strings.TrimSpace(req.NewCustomer.Name) != ""
	if hasExisting == hasNew {
		return nil, nil, orderCreateErr(ErrOnlineOrderConfirmCustomer, nil, "validation")
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	var order models.OnlineOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Items").
		First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, orderCreateErr(ErrOnlineOrderNotFound, nil, "not_found")
		}
		return nil, nil, err
	}

	if order.Status != models.OnlineOrderStatusPending {
		tx.Rollback()
		return nil, nil, orderCreateErr(ErrOnlineOrderAlreadyProcessed, nil, "already_processed")
	}
	if len(order.Items) == 0 {
		tx.Rollback()
		return nil, nil, orderCreateErr(ErrOnlineOrderValidation, nil, "validation")
	}

	var customerID uint
	if hasExisting {
		if err := validateCustomerForInvoiceTx(tx, *req.CustomerID); err != nil {
			tx.Rollback()
			return nil, nil, orderCreateErr(err, nil, "customer")
		}
		customerID = *req.CustomerID
	} else {
		customer := models.Customer{
			Name:     strings.TrimSpace(req.NewCustomer.Name),
			Phone:    strings.TrimSpace(req.NewCustomer.Phone),
			IsActive: true,
		}
		if customer.Name == "" {
			tx.Rollback()
			return nil, nil, orderCreateErr(ErrOnlineOrderConfirmCustomer, nil, "validation")
		}
		if err := tx.Create(&customer).Error; err != nil {
			tx.Rollback()
			return nil, nil, err
		}
		customerID = customer.ID
	}

	invoice := models.Invoice{
		CreatedByUserID: confirmedByUserID,
		CustomerID:      &customerID,
		Status:          models.InvoiceStatusUnpaid,
		TotalAmount:     0,
		PaidAmount:      0,
	}
	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	var totalAmount float64
	for _, item := range order.Items {
		var product models.Product
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Category").
			First(&product, item.ProductID).Error
		if err != nil {
			tx.Rollback()
			pid := item.ProductID
			return nil, nil, orderCreateErr(ErrOnlineOrderConfirmUnavailable, &pid, "unavailable")
		}
		if !product.IsActive || product.Category.ID == 0 || !product.Category.IsActive {
			tx.Rollback()
			pid := item.ProductID
			return nil, nil, orderCreateErr(ErrOnlineOrderConfirmUnavailable, &pid, "unavailable")
		}
		if product.StockQuantity < item.Quantity {
			tx.Rollback()
			pid := item.ProductID
			return nil, nil, orderCreateErr(ErrOnlineOrderConfirmStock, &pid, "insufficient_stock")
		}

		// Snapshot price — never current effectiveSalePrice.
		unitPrice := item.UnitPrice
		lineTotal := item.TotalPrice
		if lineTotal == 0 {
			lineTotal = pricing.RoundToTwoDecimals(unitPrice * item.Quantity)
		}
		totalAmount = pricing.RoundToTwoDecimals(totalAmount + lineTotal)

		invoiceItem := models.InvoiceItem{
			InvoiceID:  invoice.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: lineTotal,
		}
		if err := tx.Create(&invoiceItem).Error; err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		product.StockQuantity -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		movement := models.InventoryMovement{
			ProductID:       product.ID,
			CreatedByUserID: confirmedByUserID,
			MovementType:    "sale",
			Quantity:        item.Quantity,
			Note:            "Prodaja kroz online narudžbinu",
		}
		if err := tx.Create(&movement).Error; err != nil {
			tx.Rollback()
			return nil, nil, err
		}
	}

	invoice.TotalAmount = totalAmount
	invoice.PaidAmount = 0
	invoice.Status = models.InvoiceStatusUnpaid
	if err := tx.Save(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Model(&models.Customer{}).Where("id = ?", customerID).
		Update("total_debt", gorm.Expr("total_debt + ?", totalAmount)).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	now := time.Now().UTC()
	order.Status = models.OnlineOrderStatusConfirmed
	order.InvoiceID = &invoice.ID
	order.ConfirmedAt = &now
	order.ConfirmedByUserID = &confirmedByUserID
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	_ = database.DB.Preload("Items").First(&order, order.ID)
	_ = database.DB.Preload("Items").First(&invoice, invoice.ID)
	return &order, &invoice, nil
}

func DeletePendingOnlineOrder(orderID uint) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var order models.OnlineOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOnlineOrderNotFound
		}
		return err
	}
	if order.Status != models.OnlineOrderStatusPending {
		tx.Rollback()
		return ErrOnlineOrderDeleteNotPending
	}

	if err := tx.Where("online_order_id = ?", order.ID).Delete(&models.OnlineOrderItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&order).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
