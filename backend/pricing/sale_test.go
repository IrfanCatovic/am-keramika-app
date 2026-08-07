package pricing

import "testing"

func TestGetEffectiveSalePrice_NoSale(t *testing.T) {
	got := GetEffectiveSalePrice(2350, false, 15)
	if got != 2350 {
		t.Fatalf("want 2350 got %v", got)
	}
}

func TestGetEffectiveSalePrice_Sale15(t *testing.T) {
	// 2350 * 0.85 = 1997.50 → round up to 10 → 2000
	got := GetEffectiveSalePrice(2350, true, 15)
	if got != 2000 {
		t.Fatalf("want 2000 got %v", got)
	}
}

func TestGetEffectiveSalePrice_AlreadyOnTen(t *testing.T) {
	// 2500 * 0.8 = 2000 → already on ten → 2000
	got := GetEffectiveSalePrice(2500, true, 20)
	if got != 2000 {
		t.Fatalf("want 2000 got %v", got)
	}
}

func TestValidateSaleDiscount_Invalid(t *testing.T) {
	if err := ValidateSaleDiscount(false, -1); err != ErrNegativeDiscount {
		t.Fatalf("want ErrNegativeDiscount got %v", err)
	}
	if err := ValidateSaleDiscount(false, 100); err != ErrDiscountTooHigh {
		t.Fatalf("want ErrDiscountTooHigh got %v", err)
	}
	if err := ValidateSaleDiscount(true, 0); err != ErrSaleRequiresDiscount {
		t.Fatalf("want ErrSaleRequiresDiscount got %v", err)
	}
	if err := ValidateSaleDiscount(false, 15); err != nil {
		t.Fatalf("off-sale with stored discount should be ok: %v", err)
	}
	if err := ValidateSaleDiscount(true, 15); err != nil {
		t.Fatalf("valid sale: %v", err)
	}
}

func TestGetEffectiveSalePrice_SaleOffIgnoresDiscount(t *testing.T) {
	got := GetEffectiveSalePrice(2350, false, 50)
	if got != 2350 {
		t.Fatalf("want 2350 got %v", got)
	}
}
