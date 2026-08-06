package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCustomerHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Customer{}, &models.Invoice{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func TestGetCustomerByIDReturnsActualDebt(t *testing.T) {
	setupCustomerHandlerTestDB(t)

	customer := models.Customer{
		Name:      "Kupac A",
		Phone:     "061111111",
		TotalDebt: 250.75,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/customers/:id", GetCustomerByID)

	req := httptest.NewRequest(http.MethodGet, "/customers/"+strconv.FormatUint(uint64(customer.ID), 10), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var resp dto.CustomerDetailsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Debt != 250.75 {
		t.Fatalf("expected debt 250.75, got %v", resp.Debt)
	}
	if !resp.IsActive {
		t.Fatal("expected isActive=true for newly created customer")
	}
}
