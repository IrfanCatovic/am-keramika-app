package pricing

import (
	"errors"
	"math"
)

const QuantityEpsilon = 1e-9

var (
	ErrPackageQuantityRequired = errors.New("za prodaju po pakovanju količina u paketu je obavezna i mora biti veća od 0")
	ErrInvalidPackageQuantity  = errors.New("količina u paketu mora biti validan konačan broj veći od 0")
)

// RoundQuantity zaokružuje količinu na 4 decimale radi sprečavanja grešaka u float reprezentaciji.
func RoundQuantity(quantity float64) float64 {
	return math.Round(quantity*10000) / 10000
}

// CalculatePackagePrice računa cijenu jednog paketa na osnovu efektivne cijene po osnovnoj jedinici i količine u paketu.
// packagePrice = effectiveSalePrice × packageQuantity
func CalculatePackagePrice(effectiveSalePrice, packageQuantity float64) float64 {
	if effectiveSalePrice <= 0 || packageQuantity <= 0 {
		return 0
	}
	return RoundToTwoDecimals(effectiveSalePrice * packageQuantity)
}

// CalculatePackagedQuantity računa potreban broj punih paketa i stvarnu količinu u osnovnim jedinicama.
// packageCount = ceil(requestedQuantity / packageQuantity)
// actualQuantity = packageCount × packageQuantity
func CalculatePackagedQuantity(requestedQuantity, packageQuantity float64) (int, float64) {
	if packageQuantity <= 0 || requestedQuantity <= 0 {
		return 0, 0
	}

	ratio := requestedQuantity / packageQuantity
	nearestInt := math.Round(ratio)
	if math.Abs(ratio-nearestInt) < QuantityEpsilon {
		ratio = nearestInt
	}

	packageCount := int(math.Ceil(ratio))
	if packageCount < 0 {
		packageCount = 0
	}

	actualQuantity := RoundQuantity(float64(packageCount) * packageQuantity)
	return packageCount, actualQuantity
}

// ValidatePackageContract provjerava validnost postavki prodaje po pakovanju.
func ValidatePackageContract(saleByPackage bool, packageQuantity float64) error {
	if !saleByPackage {
		return nil
	}
	if math.IsNaN(packageQuantity) || math.IsInf(packageQuantity, 0) {
		return ErrInvalidPackageQuantity
	}
	if packageQuantity <= 0 {
		return ErrPackageQuantityRequired
	}
	return nil
}
