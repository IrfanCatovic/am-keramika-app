package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

func TestMapInvoiceResponseWithCustomer(t *testing.T) {
	customerID := uint(3)
	createdAt := time.Date(2026, 8, 6, 10, 30, 0, 0, time.UTC)

	invoice := models.Invoice{
		Model:           gormModel(1, createdAt),
		CreatedByUserID: 7,
		CreatedByUser:   models.User{Model: gormModel(7, createdAt), Username: "sef", FullName: "Šef Firma"},
		CustomerID:      &customerID,
		Customer:        &models.Customer{Model: gormModel(3, createdAt), Name: "Kupac", Phone: "061234567", IsActive: true},
		TotalAmount:     100,
		PaidAmount:      40,
		Status:          models.InvoiceStatusPartiallyPaid,
		Items: []models.InvoiceItem{
			{
				ProductID:  5,
				Product:    models.Product{Model: gormModel(5, createdAt), Name: "Pločica", Unit: "m2"},
				Quantity:   2,
				UnitPrice:  50,
				TotalPrice: 100,
			},
		},
	}

	resp := mapInvoiceResponse(invoice)

	if resp.PaidAmount != 40 || resp.RemainingAmount != 60 {
		t.Fatalf("expected paid 40 remaining 60, got paid=%v remaining=%v", resp.PaidAmount, resp.RemainingAmount)
	}
	if resp.CreatedAt != "2026-08-06 10:30" {
		t.Fatalf("unexpected createdAt: %q", resp.CreatedAt)
	}
	if resp.CustomerID == nil || *resp.CustomerID != 3 {
		t.Fatalf("expected customerID 3, got %v", resp.CustomerID)
	}
	if resp.Customer == nil || resp.Customer.Name != "Kupac" {
		t.Fatal("expected customer in response")
	}
	if !resp.Customer.IsActive {
		t.Fatal("expected customer isActive=true")
	}
	if resp.CreatedByUser == nil || resp.CreatedByUser.Username != "sef" {
		t.Fatal("expected createdByUser in response")
	}
	if resp.CreatedByUser.FullName != "Šef Firma" {
		t.Fatalf("expected fullName, got %q", resp.CreatedByUser.FullName)
	}
	if len(resp.Items) != 1 || resp.Items[0].ProductName != "Pločica" || resp.Items[0].Unit != "m2" {
		t.Fatalf("unexpected items: %+v", resp.Items)
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if containsString(string(raw), "Password") || containsString(string(raw), "passwordHash") {
		t.Fatalf("response must not expose password fields: %s", raw)
	}
}

func TestMapInvoiceResponseCashWithoutCustomer(t *testing.T) {
	createdAt := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	invoice := models.Invoice{
		Model:           gormModel(2, createdAt),
		CreatedByUserID:   1,
		CreatedByUser:     models.User{Model: gormModel(1, createdAt), Username: "radnik1"},
		CustomerID:        nil,
		Customer:          nil,
		TotalAmount:       50,
		PaidAmount:        50,
		Status:            models.InvoiceStatusPaid,
		Items:             []models.InvoiceItem{},
	}

	resp := mapInvoiceResponse(invoice)

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded["customerID"] != nil {
		t.Fatalf("expected customerID null, got %v", decoded["customerID"])
	}
	if decoded["customer"] != nil {
		t.Fatalf("expected customer null, got %v", decoded["customer"])
	}
	if resp.RemainingAmount != 0 {
		t.Fatalf("expected remaining 0, got %v", resp.RemainingAmount)
	}
}

func TestCreateInvoiceResponseUsesDTOShape(t *testing.T) {
	// Provjerava da mapInvoiceResponse daje DTO polja koja create/get dijele.
	resp := mapInvoiceResponse(models.Invoice{
		Model:       gormModel(1, time.Now()),
		TotalAmount: 10,
		PaidAmount:  10,
		Status:      models.InvoiceStatusPaid,
		Items:       []models.InvoiceItem{},
	})

	var dtoShape dto.InvoiceResponse
	raw, _ := json.Marshal(resp)
	if err := json.Unmarshal(raw, &dtoShape); err != nil {
		t.Fatalf("response is not valid InvoiceResponse DTO: %v", err)
	}
}

func gormModel(id uint, createdAt time.Time) gorm.Model {
	return gorm.Model{ID: id, CreatedAt: createdAt}
}

func containsString(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexString(s, sub) >= 0)
}

func indexString(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
