package pricing

import (
	"math"
	"testing"
)

func TestCalculateSalesReturnItemTotalStandard(t *testing.T) {
	got := CalculateSalesReturnItemTotal(2, 5000)
	if got != 10000 {
		t.Fatalf("got %v want 10000", got)
	}
}

func TestCalculateSalesReturnItemTotalPackagePhysicalQty(t *testing.T) {
	got := CalculateSalesReturnItemTotal(20, 1600)
	if got != 32000 {
		t.Fatalf("got %v want 32000", got)
	}
}

func TestPrepareSalesReturnSingleStandardItem(t *testing.T) {
	prepared, err := PrepareSalesReturn("", []SalesReturnLineInput{
		{ProductID: 1, Quantity: 2, UnitPrice: 5000},
	}, map[uint]SalesReturnProductSnapshot{
		1: {ID: 1, Name: "Bojler", Unit: "kom", SaleByPackage: false},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if prepared.TotalAmount != 10000 {
		t.Fatalf("total=%v want 10000", prepared.TotalAmount)
	}
	if len(prepared.Items) != 1 {
		t.Fatalf("items=%d want 1", len(prepared.Items))
	}
	item := prepared.Items[0]
	if item.Quantity != 2 || item.UnitPrice != 5000 || item.TotalPrice != 10000 {
		t.Fatalf("item=%+v", item)
	}
	if item.SaleByPackage || item.PackageQuantity != 0 {
		t.Fatalf("unexpected package snapshot: %+v", item)
	}
}

func TestPrepareSalesReturnPackageItemKeepsExactQuantity(t *testing.T) {
	prepared, err := PrepareSalesReturn("višak", []SalesReturnLineInput{
		{ProductID: 17, Quantity: 20, UnitPrice: 1600},
	}, map[uint]SalesReturnProductSnapshot{
		17: {
			ID:              17,
			Name:            "Calacatta",
			Unit:            "m²",
			SaleByPackage:   true,
			PackageQuantity: 1.44,
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	item := prepared.Items[0]
	if item.Quantity != 20 {
		t.Fatalf("quantity=%v want exact 20 (no package ceil)", item.Quantity)
	}
	if item.UnitPrice != 1600 || item.TotalPrice != 32000 {
		t.Fatalf("unit=%v total=%v", item.UnitPrice, item.TotalPrice)
	}
	if !item.SaleByPackage || item.PackageQuantity != 1.44 {
		t.Fatalf("package snapshot=%+v", item)
	}
	if item.ProductName != "Calacatta" || item.Unit != "m²" {
		t.Fatalf("snapshot name/unit=%+v", item)
	}

	// Explicitly assert we did NOT convert 20 m² into ceil(20/1.44)*1.44.
	_, saleActual := CalculatePackagedQuantity(20, 1.44)
	if item.Quantity == saleActual {
		t.Fatalf("return quantity must not equal sale packaged actual %v", saleActual)
	}
}

func TestPrepareSalesReturnMultipleItems(t *testing.T) {
	prepared, err := PrepareSalesReturn("Marko Marković - višak robe", []SalesReturnLineInput{
		{ProductID: 17, Quantity: 20, UnitPrice: 1600},
		{ProductID: 31, Quantity: 2, UnitPrice: 4500},
	}, map[uint]SalesReturnProductSnapshot{
		17: {ID: 17, Name: "Calacatta", Unit: "m²", SaleByPackage: true, PackageQuantity: 1.44},
		31: {ID: 31, Name: "Slavina X", Unit: "kom", SaleByPackage: false},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if prepared.Description != "Marko Marković - višak robe" {
		t.Fatalf("description=%q", prepared.Description)
	}
	if prepared.TotalAmount != 41000 {
		t.Fatalf("total=%v want 41000", prepared.TotalAmount)
	}
	if len(prepared.Items) != 2 {
		t.Fatalf("items=%d want 2", len(prepared.Items))
	}
	if prepared.Items[0].TotalPrice != 32000 || prepared.Items[1].TotalPrice != 9000 {
		t.Fatalf("line totals=%v %v", prepared.Items[0].TotalPrice, prepared.Items[1].TotalPrice)
	}
}

func TestPrepareSalesReturnRejectsInvalidQuantity(t *testing.T) {
	products := map[uint]SalesReturnProductSnapshot{
		1: {ID: 1, Name: "X", Unit: "kom"},
	}
	cases := []float64{0, -1, -0.01, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, qty := range cases {
		_, err := PrepareSalesReturn("", []SalesReturnLineInput{
			{ProductID: 1, Quantity: qty, UnitPrice: 100},
		}, products)
		if err != ErrSalesReturnInvalidQuantity {
			t.Fatalf("qty=%v want ErrSalesReturnInvalidQuantity got %v", qty, err)
		}
	}
}

func TestPrepareSalesReturnRejectsInvalidUnitPrice(t *testing.T) {
	products := map[uint]SalesReturnProductSnapshot{
		1: {ID: 1, Name: "X", Unit: "kom"},
	}
	cases := []float64{0, -1, -0.01, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, price := range cases {
		_, err := PrepareSalesReturn("", []SalesReturnLineInput{
			{ProductID: 1, Quantity: 1, UnitPrice: price},
		}, products)
		if err != ErrSalesReturnInvalidUnitPrice {
			t.Fatalf("price=%v want ErrSalesReturnInvalidUnitPrice got %v", price, err)
		}
	}
}

func TestPrepareSalesReturnRejectsEmptyItems(t *testing.T) {
	_, err := PrepareSalesReturn("note", nil, map[uint]SalesReturnProductSnapshot{})
	if err != ErrSalesReturnEmptyItems {
		t.Fatalf("want ErrSalesReturnEmptyItems got %v", err)
	}
	_, err = PrepareSalesReturn("note", []SalesReturnLineInput{}, map[uint]SalesReturnProductSnapshot{})
	if err != ErrSalesReturnEmptyItems {
		t.Fatalf("want ErrSalesReturnEmptyItems got %v", err)
	}
}

func TestPrepareSalesReturnRejectsDuplicateProductID(t *testing.T) {
	_, err := PrepareSalesReturn("", []SalesReturnLineInput{
		{ProductID: 17, Quantity: 1, UnitPrice: 100},
		{ProductID: 17, Quantity: 2, UnitPrice: 100},
	}, map[uint]SalesReturnProductSnapshot{
		17: {ID: 17, Name: "Calacatta", Unit: "m²"},
	})
	if err != ErrSalesReturnDuplicateProduct {
		t.Fatalf("want ErrSalesReturnDuplicateProduct got %v", err)
	}
}

func TestPrepareSalesReturnRejectsMissingProduct(t *testing.T) {
	_, err := PrepareSalesReturn("", []SalesReturnLineInput{
		{ProductID: 99, Quantity: 1, UnitPrice: 100},
	}, map[uint]SalesReturnProductSnapshot{})
	if err != ErrSalesReturnProductNotFound {
		t.Fatalf("want ErrSalesReturnProductNotFound got %v", err)
	}
}

func TestPrepareSalesReturnDoesNotCeilPackageQuantity(t *testing.T) {
	// Sale path for 10 m² / 1.44 would become 10.08; return must stay 10.
	prepared, err := PrepareSalesReturn("", []SalesReturnLineInput{
		{ProductID: 17, Quantity: 10, UnitPrice: 1600},
	}, map[uint]SalesReturnProductSnapshot{
		17: {ID: 17, Name: "Calacatta", Unit: "m²", SaleByPackage: true, PackageQuantity: 1.44},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if prepared.Items[0].Quantity != 10 {
		t.Fatalf("quantity=%v want 10", prepared.Items[0].Quantity)
	}
	if prepared.Items[0].TotalPrice != 16000 {
		t.Fatalf("total=%v want 16000", prepared.Items[0].TotalPrice)
	}
}
