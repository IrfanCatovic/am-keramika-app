package repositories

import (
	"errors"
	"math"
	"net/mail"
	"strings"
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
