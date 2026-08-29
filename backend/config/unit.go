package config

import (
	"errors"
	"strings"
)

const (
	UnitSquareMeter = "m²"
	UnitPiece       = "kom"
)

var ErrInvalidProductUnit = errors.New("jedinica mora biti m² (metar kvadratni) ili kom (komad)")

// NormalizeProductUnit converts legacy unit values to the two supported
// product units. Unknown values are returned trimmed so callers can validate
// them without silently changing their meaning.
func NormalizeProductUnit(value string) string {
	unit := strings.TrimSpace(value)
	normalized := strings.ToLower(strings.Join(strings.Fields(unit), ""))

	switch normalized {
	case "m2", "m²", "m^2", "kv", "kvadrat", "kvadrata":
		return UnitSquareMeter
	case "kom", "komad", "komada", "pcs", "piece":
		return UnitPiece
	default:
		return unit
	}
}

func ValidateProductUnit(value string) error {
	switch NormalizeProductUnit(value) {
	case UnitSquareMeter, UnitPiece:
		return nil
	default:
		return ErrInvalidProductUnit
	}
}
