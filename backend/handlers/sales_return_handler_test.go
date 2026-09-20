package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func setupSalesReturnHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.InventoryMovement{},
		&models.Refund{},
		&models.SalesReturn{},
		&models.SalesReturnItem{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-sales-returns")
}

func setupSalesReturnHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/sales-returns", handlers.CreateSalesReturn)
		}
	}
	return r
}

func salesReturnLogin(t *testing.T, r *gin.Engine, username, role string) (string, models.User) {
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
	return token, user
}

func TestCreateSalesReturnWorkerCanExecute(t *testing.T) {
	setupSalesReturnHandlerTestDB(t)
	r := setupSalesReturnHandlerRouter()
	token, user := salesReturnLogin(t, r, "radnik1", models.RoleWorker)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Bojler", Slug: "bojler", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 5000, StockQuantity: 5, IsActive: true,
	}
	database.DB.Create(&product)

	body, _ := json.Marshal(map[string]any{
		"description":  "Marko Marković - višak materijala",
		"cashRefunded": true,
		"items": []map[string]any{
			{"productID": product.ID, "quantity": 2, "unitPrice": 5000},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/sales-returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp dto.SalesReturnResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalAmount != 10000 || !resp.CashRefunded || resp.RefundID == nil {
		t.Fatalf("response=%+v", resp)
	}
	if resp.CreatedByUser == nil || resp.CreatedByUser.ID != user.ID {
		t.Fatalf("createdBy=%+v want user %d", resp.CreatedByUser, user.ID)
	}
}

func TestCreateSalesReturnRejectsMissingCashRefunded(t *testing.T) {
	setupSalesReturnHandlerTestDB(t)
	r := setupSalesReturnHandlerRouter()
	token, _ := salesReturnLogin(t, r, "radnik1", models.RoleWorker)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Bojler", Slug: "bojler", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 5000, StockQuantity: 5, IsActive: true,
	}
	database.DB.Create(&product)

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{
			{"productID": product.ID, "quantity": 2, "unitPrice": 5000},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/sales-returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
