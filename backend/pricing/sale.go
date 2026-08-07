package pricing

import "errors"

var (
	ErrNegativeDiscount     = errors.New("popust ne sme biti negativan")
	ErrDiscountTooHigh      = errors.New("popust mora biti manji od 100%")
	ErrSaleRequiresDiscount = errors.New("za akciju popust mora biti veći od 0%")
)

// GetEffectiveSalePrice vraća cijenu koju kupac trenutno plaća.
// Regularni salePrice se ne mijenja; akcija je derived.
func GetEffectiveSalePrice(salePrice float64, isOnSale bool, discountPercent float64) float64 {
	if !isOnSale || discountPercent <= 0 {
		return salePrice
	}
	raw := salePrice * (1 - discountPercent/100)
	return RoundUpToTen(raw)
}

// ValidateSaleDiscount provjerava konzistentnost isOnSale + discountPercent.
func ValidateSaleDiscount(isOnSale bool, discountPercent float64) error {
	if discountPercent < 0 {
		return ErrNegativeDiscount
	}
	if discountPercent >= 100 {
		return ErrDiscountTooHigh
	}
	if isOnSale && discountPercent <= 0 {
		return ErrSaleRequiresDiscount
	}
	return nil
}

// DiscountedRawSalePrice is the pre-round amount for UI preview.
func DiscountedRawSalePrice(salePrice, discountPercent float64) float64 {
	return RoundToTwoDecimals(salePrice * (1 - discountPercent/100))
}
