package pricing

import (
	"math"
	"testing"
)

func TestCalculatePackagePrice(t *testing.T) {
	// Effective 2000 × 2.70 = 5400
	got := CalculatePackagePrice(2000, 2.70)
	if got != 5400 {
		t.Fatalf("CalculatePackagePrice(2000, 2.70) want 5400, got %v", got)
	}

	// Effective with sale: regular 2350, 15% discount -> effective 2000, package 2.70 -> 5400
	effective := GetEffectiveSalePrice(2350, true, 15)
	if effective != 2000 {
		t.Fatalf("GetEffectiveSalePrice(2350, true, 15) want 2000, got %v", effective)
	}
	gotSale := CalculatePackagePrice(effective, 2.70)
	if gotSale != 5400 {
		t.Fatalf("CalculatePackagePrice(effective 2000, 2.70) want 5400, got %v", gotSale)
	}

	// Decimal package quantities: e.g. 1.44 m² × 1850 RSD
	gotDecimal := CalculatePackagePrice(1850, 1.44)
	if gotDecimal != 2664 {
		t.Fatalf("CalculatePackagePrice(1850, 1.44) want 2664, got %v", gotDecimal)
	}

	// Zero / negative
	if CalculatePackagePrice(0, 2.70) != 0 {
		t.Fatalf("want 0 for 0 price")
	}
	if CalculatePackagePrice(2000, 0) != 0 {
		t.Fatalf("want 0 for 0 package quantity")
	}
}

func TestCalculatePackagedQuantity(t *testing.T) {
	// 7. requested 10, package 2.70 -> packageCount 4, actualQuantity 10.80
	count, actual := CalculatePackagedQuantity(10, 2.70)
	if count != 4 {
		t.Fatalf("packageCount want 4, got %v", count)
	}
	if actual != 10.80 {
		t.Fatalf("actualQuantity want 10.80, got %v", actual)
	}

	// 8. requested exact full package: requested 10.80, package 2.70 -> 4 packages, 10.80
	count8, actual8 := CalculatePackagedQuantity(10.80, 2.70)
	if count8 != 4 {
		t.Fatalf("packageCount want 4, got %v", count8)
	}
	if actual8 != 10.80 {
		t.Fatalf("actualQuantity want 10.80, got %v", actual8)
	}

	// 9. slightly over: requested 10.81, package 2.70 -> 5 packages, 13.50
	count9, actual9 := CalculatePackagedQuantity(10.81, 2.70)
	if count9 != 5 {
		t.Fatalf("packageCount want 5, got %v", count9)
	}
	if actual9 != 13.50 {
		t.Fatalf("actualQuantity want 13.50, got %v", actual9)
	}

	// Various package sizes: 1.44
	count144, actual144 := CalculatePackagedQuantity(5, 1.44)
	// ceil(5 / 1.44) = ceil(3.4722) = 4 -> 4 * 1.44 = 5.76
	if count144 != 4 || actual144 != 5.76 {
		t.Fatalf("want 4 packages and 5.76 actual, got %v and %v", count144, actual144)
	}

	// Zero or negative input
	c0, a0 := CalculatePackagedQuantity(0, 2.70)
	if c0 != 0 || a0 != 0 {
		t.Fatalf("want 0, 0 for requested 0")
	}
}

func TestValidatePackageContract(t *testing.T) {
	// saleByPackage = false -> valid regardless of quantity
	if err := ValidatePackageContract(false, 0); err != nil {
		t.Fatalf("unexpected error for saleByPackage=false: %v", err)
	}
	if err := ValidatePackageContract(false, 2.70); err != nil {
		t.Fatalf("unexpected error for saleByPackage=false: %v", err)
	}

	// saleByPackage = true, valid quantity
	if err := ValidatePackageContract(true, 2.70); err != nil {
		t.Fatalf("unexpected error for valid package: %v", err)
	}

	// saleByPackage = true, quantity 0 -> error
	if err := ValidatePackageContract(true, 0); err == nil {
		t.Fatalf("expected error for quantity 0")
	}

	// saleByPackage = true, negative quantity -> error
	if err := ValidatePackageContract(true, -1.5); err == nil {
		t.Fatalf("expected error for negative quantity")
	}

	// saleByPackage = true, NaN or Inf -> error
	if err := ValidatePackageContract(true, math.NaN()); err == nil {
		t.Fatalf("expected error for NaN quantity")
	}
	if err := ValidatePackageContract(true, math.Inf(1)); err == nil {
		t.Fatalf("expected error for Inf quantity")
	}
}
