package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
)

func postAvailability(t *testing.T, r http.Handler, productID uint, quantity float64) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(dto.PublicAvailabilityCheckRequest{Quantity: quantity})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/public/products/"+strconv.FormatUint(uint64(productID), 10)+"/check-availability",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	return response
}

func TestPublicAvailabilityUsesActualPackageQuantity(t *testing.T) {
	setupPricingHandlerTestDB(t)
	r := setupSaleRouter()
	_, _, _, product, _ := seedPublicCatalog(t)
	product.SaleByPackage = true
	product.PackageQuantity = 2.70
	product.StockQuantity = 10.80
	if err := database.DB.Save(&product).Error; err != nil {
		t.Fatalf("save product: %v", err)
	}

	w := postAvailability(t, r, product.ID, 10)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 got %d: %s", w.Code, w.Body.String())
	}
	var response dto.PublicAvailabilityCheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Available || response.PackageCount != 4 ||
		response.ActualQuantity == nil || *response.ActualQuantity != 10.80 {
		t.Fatalf("unexpected response: %+v", response)
	}

	product.StockQuantity = 10.70
	if err := database.DB.Save(&product).Error; err != nil {
		t.Fatalf("save reduced stock: %v", err)
	}
	w = postAvailability(t, r, product.ID, 10)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200 got %d: %s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode insufficient response: %v", err)
	}
	if response.Available {
		t.Fatalf("package availability should reject leftover stock: %+v", response)
	}
}
