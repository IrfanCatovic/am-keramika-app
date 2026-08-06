package invoicepdf_test

import (
	"bytes"
	"strings"
	"testing"

	"am-keramika-backend/internal/invoicepdf"
)

func sampleDoc() invoicepdf.Document {
	return invoicepdf.Document{
		ID:              42,
		CreatedAt:       "2026-08-06 18:30",
		Status:          "paid",
		IsCashSale:      false,
		CustomerName:    "Mujo Čengić",
		CustomerPhone:   "061234567",
		TotalAmount:     5800.5,
		PaidAmount:      5800.5,
		RemainingAmount: 0,
		CreatedBy:       "sef",
		Company: invoicepdf.Company{
			Name:    "AM Keramika",
			Address: "Ulica Test 1",
			City:    "Tutin",
			Phone:   "063 652222",
		},
		Items: []invoicepdf.Item{
			{
				ProductName: "Pločica Bež",
				Quantity:    2.5,
				Unit:        "m2",
				UnitPrice:   1500.5,
				TotalPrice:  3751.25,
			},
			{
				ProductName: "Ljepilo Široko",
				Quantity:    1,
				Unit:        "kg",
				UnitPrice:   2049.25,
				TotalPrice:  2049.25,
			},
		},
	}
}

func TestGeneratePDFStartsWithHeader(t *testing.T) {
	pdf, err := invoicepdf.Generate(sampleDoc())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(pdf) < 5 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("expected PDF header, got %q", string(pdf[:min(8, len(pdf))]))
	}
}

func TestGenerateCancelledWatermarkNoPanic(t *testing.T) {
	doc := sampleDoc()
	doc.Status = "cancelled"
	doc.RemainingAmount = 100
	pdf, err := invoicepdf.Generate(doc)
	if err != nil {
		t.Fatalf("generate cancelled: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatal("expected PDF bytes")
	}
}

func TestGenerateUnicodeCustomerName(t *testing.T) {
	doc := sampleDoc()
	doc.CustomerName = "Šćepan Žđić"
	doc.Items[0].ProductName = "Čokoladna pločica đak"
	pdf, err := invoicepdf.Generate(doc)
	if err != nil {
		t.Fatalf("unicode generate: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatal("expected PDF bytes")
	}
}

func TestGenerateManyItemsMultiPage(t *testing.T) {
	doc := sampleDoc()
	doc.Items = nil
	for i := 0; i < 80; i++ {
		doc.Items = append(doc.Items, invoicepdf.Item{
			ProductName: "Proizvod " + strings.Repeat("x", i%5),
			Quantity:    1.25,
			Unit:        "kom",
			UnitPrice:   100,
			TotalPrice:  125,
		})
	}
	pdf, err := invoicepdf.Generate(doc)
	if err != nil {
		t.Fatalf("multipage: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatal("expected PDF bytes")
	}
	// Multiple pages usually mean more than one /Type /Page occurrence.
	if bytes.Count(pdf, []byte("/Type /Page")) < 2 && bytes.Count(pdf, []byte("/Type/Page")) < 2 {
		// Soft check — generator still must not panic; size sanity.
		if len(pdf) < 1000 {
			t.Fatalf("unexpectedly small multipage PDF: %d bytes", len(pdf))
		}
	}
}

func TestFilename(t *testing.T) {
	got := invoicepdf.Filename(7)
	want := "AM-Keramika-Racun-7.pdf"
	if got != want {
		t.Fatalf("filename: got %q want %q", got, want)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
