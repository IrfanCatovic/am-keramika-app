package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRefundHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Customer{},
		&models.Invoice{},
		&models.Refund{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-refunds")
}

func setupRefundHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		finance := authorized.Group("/")
		finance.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager))
		{
			finance.GET("/refunds", handlers.GetRefunds)
		}
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleWorker))
		{
			staff.GET("/refunds-worker-probe", handlers.GetRefunds)
		}
	}
	return r
}

func refundLogin(t *testing.T, r *gin.Engine, username, role string) string {
	t.Helper()
	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: username, PasswordHash: hash, Role: role, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"username": username, "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, _ := resp["token"].(string)
	return token
}

func TestGetRefundsStaffAuthAndList(t *testing.T) {
	setupRefundHandlerTestDB(t)
	r := setupRefundHandlerRouter()
	token := refundLogin(t, r, "boss1", models.RoleBoss)

	customer := models.Customer{Name: "Kupac", Phone: "1", IsActive: true}
	database.DB.Create(&customer)
	var user models.User
	database.DB.Where("username = ?", "boss1").First(&user)
	invoice := models.Invoice{
		CreatedByUserID: user.ID,
		CustomerID:      &customer.ID,
		TotalAmount:     2000,
		PaidAmount:      2000,
		Status:          models.InvoiceStatusCancelled,
	}
	database.DB.Create(&invoice)
	database.DB.Create(&models.Refund{
		InvoiceID: invoice.ID, CreatedByUserID: user.ID, Amount: 2000, Reason: "Storno",
	})

	req := httptest.NewRequest(http.MethodGet, "/refunds?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp dto.PaginatedRefundsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Refunds) != 1 || resp.Refunds[0].Amount != 2000 {
		t.Fatalf("unexpected refunds %+v", resp.Refunds)
	}

	req2 := httptest.NewRequest(
		http.MethodGet,
		"/refunds?invoiceID="+strconv.FormatUint(uint64(invoice.ID), 10),
		nil,
	)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("invoice filter status=%d", w2.Code)
	}
}

func TestGetRefundsWorkerForbiddenOnFinanceRoute(t *testing.T) {
	setupRefundHandlerTestDB(t)
	r := setupRefundHandlerRouter()
	token := refundLogin(t, r, "worker1", models.RoleWorker)

	req := httptest.NewRequest(http.MethodGet, "/refunds", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on /refunds, got %d", w.Code)
	}
}
