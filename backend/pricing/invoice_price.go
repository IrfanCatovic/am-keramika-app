package pricing

import (
	"errors"
	"math"
)

var ErrInvalidPriceOverride = errors.New("ručna cena mora biti veća od 0")

// ResolveInvoiceItemUnitPrice bira konačnu cenu stavke računa.
// Prioritet: priceOverride (ako postoji) > effectiveSalePrice.
// Ručna cena je konačna — DiscountPercent se ne primenjuje preko nje.
func ResolveInvoiceItemUnitPrice(
	effectiveSalePrice float64,
	priceOverride *float64,
) (unitPrice float64, originalUnitPrice float64, overridden bool, err error) {
	originalUnitPrice = effectiveSalePrice
	if priceOverride == nil {
		return effectiveSalePrice, originalUnitPrice, false, nil
	}

	override := *priceOverride
	if math.IsNaN(override) || math.IsInf(override, 0) || override <= 0 {
		return 0, 0, false, ErrInvalidPriceOverride
	}

	unitPrice = RoundToTwoDecimals(override)
	if unitPrice <= 0 {
		return 0, 0, false, ErrInvalidPriceOverride
	}
	return unitPrice, originalUnitPrice, true, nil
}
