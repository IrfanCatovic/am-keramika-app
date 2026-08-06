package pricing

import (
	"errors"
	"math"
)

const (
	ModeManual     = "manual"
	ModeCalculated = "calculated"
)

var (
	ErrNegativePurchasePrice = errors.New("nabavna cijena ne smije biti negativna")
	ErrNegativeSalePrice     = errors.New("prodajna cijena ne smije biti negativna")
	ErrNegativeMargin        = errors.New("marža ne smije biti negativna")
	ErrNegativeVAT           = errors.New("PDV ne smije biti negativan")
	ErrPurchaseRequired      = errors.New("za automatski obračun nabavna cijena je obavezna i mora biti veća od 0")
	ErrManualSaleRequired    = errors.New("za ručni način prodajna cijena je obavezna i mora biti veća od 0")
)

// Input su ulazne vrijednosti za obračun cijene.
type Input struct {
	PurchasePrice *float64
	MarginPercent *float64
	VatPercent    *float64
	SalePrice     *float64 // tražena ručna cijena; u calculated mode se ignoriše
}

// Result je rezultat obračuna.
type Result struct {
	Mode             string
	PurchasePrice    *float64
	MarginPercent    *float64
	VatPercent       *float64
	PriceAfterMargin float64
	RawSalePrice     float64
	FinalSalePrice   float64
}

func NormalizeOptionalAmount(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func IsCalculatedMode(marginPercent, vatPercent float64) bool {
	return marginPercent > 0 || vatPercent > 0
}

func DetectMode(purchasePrice, marginPercent, vatPercent *float64) string {
	_ = purchasePrice
	if IsCalculatedMode(NormalizeOptionalAmount(marginPercent), NormalizeOptionalAmount(vatPercent)) {
		return ModeCalculated
	}
	return ModeManual
}

// RoundToTwoDecimals zaokružuje na 2 decimale (bankarski round half away via math.Round).
func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// RoundUpToTen zaokružuje naviše na sljedećih 10 RSD.
// Ako je vrijednost već djeljiva sa 10, ostaje ista.
func RoundUpToTen(raw float64) float64 {
	twoDecimals := RoundToTwoDecimals(raw)
	remainder := math.Mod(twoDecimals, 10)
	if remainder < 1e-9 || remainder > 10-1e-9 {
		return RoundToTwoDecimals(twoDecimals)
	}
	return math.Ceil(twoDecimals/10) * 10
}

func pointerOrNil(value float64, keepZero bool) *float64 {
	if !keepZero && value == 0 {
		return nil
	}
	v := value
	return &v
}

// Calculate primjenjuje poslovna pravila ručne i automatske cijene.
func Calculate(in Input) (Result, error) {
	purchase := NormalizeOptionalAmount(in.PurchasePrice)
	margin := NormalizeOptionalAmount(in.MarginPercent)
	vat := NormalizeOptionalAmount(in.VatPercent)
	sale := NormalizeOptionalAmount(in.SalePrice)

	if in.PurchasePrice != nil && *in.PurchasePrice < 0 {
		return Result{}, ErrNegativePurchasePrice
	}
	if in.SalePrice != nil && *in.SalePrice < 0 {
		return Result{}, ErrNegativeSalePrice
	}
	if in.MarginPercent != nil && *in.MarginPercent < 0 {
		return Result{}, ErrNegativeMargin
	}
	if in.VatPercent != nil && *in.VatPercent < 0 {
		return Result{}, ErrNegativeVAT
	}

	result := Result{
		PurchasePrice: in.PurchasePrice,
		MarginPercent: in.MarginPercent,
		VatPercent:    in.VatPercent,
	}

	if IsCalculatedMode(margin, vat) {
		if purchase <= 0 {
			return Result{}, ErrPurchaseRequired
		}
		priceAfterMargin := purchase * (1 + margin/100)
		rawSalePrice := RoundToTwoDecimals(priceAfterMargin * (1 + vat/100))
		final := RoundUpToTen(rawSalePrice)

		result.Mode = ModeCalculated
		result.PriceAfterMargin = RoundToTwoDecimals(priceAfterMargin)
		result.RawSalePrice = rawSalePrice
		result.FinalSalePrice = final
		result.PurchasePrice = pointerOrNil(purchase, true)
		if margin == 0 {
			result.MarginPercent = pointerOrNil(0, true)
		} else {
			result.MarginPercent = pointerOrNil(margin, true)
		}
		if vat == 0 {
			result.VatPercent = pointerOrNil(0, true)
		} else {
			result.VatPercent = pointerOrNil(vat, true)
		}
		return result, nil
	}

	if in.SalePrice == nil || sale <= 0 {
		return Result{}, ErrManualSaleRequired
	}

	result.Mode = ModeManual
	result.FinalSalePrice = sale
	result.RawSalePrice = sale
	result.PriceAfterMargin = sale
	if purchase > 0 {
		result.PurchasePrice = pointerOrNil(purchase, true)
	} else if in.PurchasePrice != nil && *in.PurchasePrice == 0 {
		zero := 0.0
		result.PurchasePrice = &zero
	} else {
		result.PurchasePrice = nil
	}
	zero := 0.0
	result.MarginPercent = &zero
	result.VatPercent = &zero
	return result, nil
}
