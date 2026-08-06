package pricing

import "testing"

func floatPtr(v float64) *float64 { return &v }

func TestManualSalePriceKept(t *testing.T) {
	res, err := Calculate(Input{
		MarginPercent: floatPtr(0),
		VatPercent:    floatPtr(0),
		SalePrice:     floatPtr(153),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Mode != ModeManual {
		t.Fatalf("mode=%s", res.Mode)
	}
	if res.FinalSalePrice != 153 {
		t.Fatalf("sale=%v", res.FinalSalePrice)
	}
}

func TestCalculated100_25_20(t *testing.T) {
	res, err := Calculate(Input{
		PurchasePrice: floatPtr(100),
		MarginPercent: floatPtr(25),
		VatPercent:    floatPtr(20),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.FinalSalePrice != 150 {
		t.Fatalf("want 150 got %v (raw=%v)", res.FinalSalePrice, res.RawSalePrice)
	}
}

func TestCalculated100_10_10(t *testing.T) {
	res, err := Calculate(Input{
		PurchasePrice: floatPtr(100),
		MarginPercent: floatPtr(10),
		VatPercent:    floatPtr(10),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.RawSalePrice != 121 {
		t.Fatalf("raw want 121 got %v", res.RawSalePrice)
	}
	if res.FinalSalePrice != 130 {
		t.Fatalf("final want 130 got %v", res.FinalSalePrice)
	}
}

func TestRoundUpCases(t *testing.T) {
	cases := []struct {
		raw  float64
		want float64
	}{
		{152, 160},
		{156, 160},
		{160, 160},
		{161, 170},
	}
	for _, tc := range cases {
		if got := RoundUpToTen(tc.raw); got != tc.want {
			t.Fatalf("RoundUpToTen(%v)=%v want %v", tc.raw, got, tc.want)
		}
	}
}

func TestCalculatedRequiresPurchase(t *testing.T) {
	_, err := Calculate(Input{
		MarginPercent: floatPtr(10),
		VatPercent:    floatPtr(0),
	})
	if err != ErrPurchaseRequired {
		t.Fatalf("want ErrPurchaseRequired got %v", err)
	}
}

func TestNegativeValidation(t *testing.T) {
	if _, err := Calculate(Input{PurchasePrice: floatPtr(-1), SalePrice: floatPtr(10)}); err != ErrNegativePurchasePrice {
		t.Fatalf("purchase: %v", err)
	}
	if _, err := Calculate(Input{SalePrice: floatPtr(-5)}); err != ErrNegativeSalePrice {
		t.Fatalf("sale: %v", err)
	}
	if _, err := Calculate(Input{MarginPercent: floatPtr(-1), SalePrice: floatPtr(10)}); err != ErrNegativeMargin {
		t.Fatalf("margin: %v", err)
	}
	if _, err := Calculate(Input{VatPercent: floatPtr(-2), SalePrice: floatPtr(10)}); err != ErrNegativeVAT {
		t.Fatalf("vat: %v", err)
	}
}

func TestCalculatedIgnoresRequestSalePrice(t *testing.T) {
	res, err := Calculate(Input{
		PurchasePrice: floatPtr(100),
		MarginPercent: floatPtr(25),
		VatPercent:    floatPtr(20),
		SalePrice:     floatPtr(999),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.FinalSalePrice != 150 {
		t.Fatalf("want 150 got %v", res.FinalSalePrice)
	}
}

func TestDetectMode(t *testing.T) {
	if DetectMode(nil, floatPtr(0), floatPtr(0)) != ModeManual {
		t.Fatal("expected manual")
	}
	if DetectMode(floatPtr(10), floatPtr(5), floatPtr(0)) != ModeCalculated {
		t.Fatal("expected calculated")
	}
}
