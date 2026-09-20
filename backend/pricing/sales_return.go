package pricing

import (
	"errors"
	"math"
)

var (
	ErrSalesReturnEmptyItems           = errors.New("povrat mora imati najmanje jednu stavku")
	ErrSalesReturnDuplicateProduct     = errors.New("isti proizvod ne sme biti više puta u istom povratu")
	ErrSalesReturnInvalidQuantity      = errors.New("količina mora biti veća od 0")
	ErrSalesReturnInvalidUnitPrice     = errors.New("cena povrata mora biti veća od 0")
	ErrSalesReturnProductNotFound      = errors.New("proizvod nije pronađen")
	ErrSalesReturnCashRefundedRequired = errors.New("mora se navesti da li je novac vraćen kupcu")
)

// SalesReturnProductSnapshot holds product fields needed to build return line snapshots.
type SalesReturnProductSnapshot struct {
	ID              uint
	Name            string
	Unit            string
	SaleByPackage   bool
	PackageQuantity float64
}

// SalesReturnLineInput is one requested return line before validation/calculation.
type SalesReturnLineInput struct {
	ProductID uint
	Quantity  float64
	UnitPrice float64
}

// PreparedSalesReturnItem is a validated return line ready for persistence.
// Quantity is the exact physical quantity received — never ceil'd to packages.
type PreparedSalesReturnItem struct {
	ProductID       uint
	ProductName     string
	Unit            string
	Quantity        float64
	UnitPrice       float64
	TotalPrice      float64
	SaleByPackage   bool
	PackageQuantity float64
}

// PreparedSalesReturn is a validated multi-item return ready for persistence.
type PreparedSalesReturn struct {
	Description string
	TotalAmount float64
	Items       []PreparedSalesReturnItem
}

// CalculateSalesReturnItemTotal = Quantity × UnitPrice, rounded to money precision.
// Does not apply package ceil / round-up.
func CalculateSalesReturnItemTotal(quantity, unitPrice float64) float64 {
	return RoundToTwoDecimals(quantity * unitPrice)
}

// PrepareSalesReturn validates request lines against product snapshots and
// builds immutable return item snapshots + TotalAmount.
//
// Rules:
//   - at least one item
//   - duplicate ProductID rejected
//   - quantity and unitPrice must be finite and > 0
//   - product must exist in productsByID (for name/unit/package snapshot)
//   - return Quantity stays the exact physical quantity (no package ceil)
func PrepareSalesReturn(
	description string,
	lines []SalesReturnLineInput,
	productsByID map[uint]SalesReturnProductSnapshot,
) (*PreparedSalesReturn, error) {
	if len(lines) == 0 {
		return nil, ErrSalesReturnEmptyItems
	}

	seen := make(map[uint]struct{}, len(lines))
	items := make([]PreparedSalesReturnItem, 0, len(lines))
	totalAmount := 0.0

	for _, line := range lines {
		if _, exists := seen[line.ProductID]; exists {
			return nil, ErrSalesReturnDuplicateProduct
		}
		seen[line.ProductID] = struct{}{}

		if !isFinitePositive(line.Quantity) {
			return nil, ErrSalesReturnInvalidQuantity
		}
		if !isFinitePositive(line.UnitPrice) {
			return nil, ErrSalesReturnInvalidUnitPrice
		}

		product, ok := productsByID[line.ProductID]
		if !ok || product.ID == 0 {
			return nil, ErrSalesReturnProductNotFound
		}

		quantity := RoundQuantity(line.Quantity)
		if quantity <= 0 {
			return nil, ErrSalesReturnInvalidQuantity
		}

		unitPrice := RoundToTwoDecimals(line.UnitPrice)
		if unitPrice <= 0 {
			return nil, ErrSalesReturnInvalidUnitPrice
		}

		totalPrice := CalculateSalesReturnItemTotal(quantity, unitPrice)
		packageQuantity := 0.0
		if product.SaleByPackage {
			packageQuantity = RoundQuantity(product.PackageQuantity)
		}

		items = append(items, PreparedSalesReturnItem{
			ProductID:       product.ID,
			ProductName:     product.Name,
			Unit:            product.Unit,
			Quantity:        quantity,
			UnitPrice:       unitPrice,
			TotalPrice:      totalPrice,
			SaleByPackage:   product.SaleByPackage,
			PackageQuantity: packageQuantity,
		})
		totalAmount = RoundToTwoDecimals(totalAmount + totalPrice)
	}

	return &PreparedSalesReturn{
		Description: description,
		TotalAmount: totalAmount,
		Items:       items,
	}, nil
}

func isFinitePositive(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}
