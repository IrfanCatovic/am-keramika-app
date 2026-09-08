package pricing

import (
	"math"
	"testing"
)

func TestResolveInvoiceItemUnitPriceWithoutOverride(t *testing.T) {
	unit, original, overridden, err := ResolveInvoiceItemUnitPrice(1800, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if unit != 1800 || original != 1800 || overridden {
		t.Fatalf("got unit=%v original=%v overridden=%v", unit, original, overridden)
	}
}

func TestResolveInvoiceItemUnitPriceWithOverride(t *testing.T) {
	override := 1600.0
	unit, original, overridden, err := ResolveInvoiceItemUnitPrice(1800, &override)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if unit != 1600 || original != 1800 || !overridden {
		t.Fatalf("got unit=%v original=%v overridden=%v", unit, original, overridden)
	}
}

func TestResolveInvoiceItemUnitPriceRejectsInvalid(t *testing.T) {
	cases := []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, raw := range cases {
		v := raw
		_, _, _, err := ResolveInvoiceItemUnitPrice(1000, &v)
		if err != ErrInvalidPriceOverride {
			t.Fatalf("override=%v want ErrInvalidPriceOverride got %v", raw, err)
		}
	}
}
