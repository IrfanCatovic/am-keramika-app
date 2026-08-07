package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
)

func TestCheckPublicProductAvailability(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	_, _, _, featured, regular := seedPublicCatalog(t)

	post := func(id uint, quantity float64) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(dto.PublicAvailabilityCheckRequest{Quantity: quantity})
		req := httptest.NewRequest(
			http.MethodPost,
			"/public/products/"+strconv.FormatUint(uint64(id), 10)+"/check-availability",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("active sufficient stock", func(t *testing.T) {
		w := post(featured.ID, 10)
		if w.Code != http.StatusOK {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var resp dto.PublicAvailabilityCheckResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if !resp.Available {
			t.Fatalf("want available true: %+v", resp)
		}
		raw := w.Body.String()
		if strings.Contains(raw, "stockQuantity") || strings.Contains(raw, "availableQuantity") {
			t.Fatalf("leaked stock fields: %s", raw)
		}
	})

	t.Run("active equal stock", func(t *testing.T) {
		w := post(featured.ID, 10)
		var resp dto.PublicAvailabilityCheckResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if !resp.Available {
			t.Fatalf("equal stock should be available")
		}
	})

	t.Run("quantity greater than stock", func(t *testing.T) {
		w := post(featured.ID, 10.01)
		if w.Code != http.StatusOK {
			t.Fatalf("%d", w.Code)
		}
		var resp dto.PublicAvailabilityCheckResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Available || resp.Reason != "insufficient_stock" {
			t.Fatalf("got %+v", resp)
		}
	})

	t.Run("stock zero", func(t *testing.T) {
		w := post(regular.ID, 1)
		var resp dto.PublicAvailabilityCheckResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Available {
			t.Fatalf("zero stock should be unavailable")
		}
	})

	t.Run("zero quantity", func(t *testing.T) {
		w := post(featured.ID, 0)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})

	t.Run("negative quantity", func(t *testing.T) {
		w := post(featured.ID, -1)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})

	t.Run("inactive product", func(t *testing.T) {
		var inactive models.Product
		database.DB.Where("slug = ?", "sakriven").First(&inactive)
		w := post(inactive.ID, 1)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 got %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("inactive category", func(t *testing.T) {
		var orphan models.Product
		database.DB.Where("slug = ?", "u-staroj").First(&orphan)
		w := post(orphan.ID, 1)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 got %d", w.Code)
		}
	})

	t.Run("does not mutate stock", func(t *testing.T) {
		var before models.Product
		database.DB.First(&before, featured.ID)
		post(featured.ID, 3)
		var after models.Product
		database.DB.First(&after, featured.ID)
		if after.StockQuantity != before.StockQuantity {
			t.Fatalf("stock mutated %v -> %v", before.StockQuantity, after.StockQuantity)
		}
	})
}
