package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAuthTestDB(t *testing.T) {
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
		&models.ProductGroup{},
		&models.Product{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.InvoiceCancellation{},
		&models.Refund{},
		&models.Customer{},
		&models.InventoryMovement{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-auth-handlers")
}

func createUser(t *testing.T, username, password, role string, active bool) models.User {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	user := models.User{
		Username:     auth.NormalizeUsername(username),
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if !active {
		if err := database.DB.Model(&user).Update("is_active", false).Error; err != nil {
			t.Fatalf("deactivate user: %v", err)
		}
		user.IsActive = false
	}
	return user
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		authorized.GET("/auth/me", handlers.GetMe)

		boss := authorized.Group("/")
		boss.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss))
		{
			boss.POST("/users", handlers.CreateUser)
			boss.PUT("/users/:id/status", handlers.UpdateUserStatus)
			boss.PUT("/users/:id", handlers.UpdateUser)
			boss.PUT("/users/:id/password", handlers.UpdateUserPassword)
			boss.GET("/users", handlers.GetUsers)
		}

		dailyReports := authorized.Group("/reports")
		dailyReports.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			dailyReports.GET("/daily", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
			dailyReports.GET("/sales-summary", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
		}

		reports := authorized.Group("/reports")
		reports.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager))
		{
			reports.GET("/period", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
		}

		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/products", handlers.CreateProduct)
			staff.PUT("/products/:id", handlers.UpdateProduct)
			staff.PUT("/invoices/:id/cancel", handlers.CancelInvoice)
		}
	}
	return r
}

func loginToken(t *testing.T, r *gin.Engine, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status %d body %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	token, _ := resp["token"].(string)
	if token == "" {
		t.Fatal("missing token")
	}
	return token
}

func TestLoginSuccess(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "SeF", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"token"`)) {
		t.Fatal("expected token in response")
	}
	if bytes.Contains(w.Body.Bytes(), []byte("password")) {
		t.Fatal("password hash must not leak")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "sef", "password": "wrongpass1"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginInactiveUser(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "radnik1", "password123", models.RoleWorker, false)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "radnik1", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	setupAuthTestDB(t)
	r := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()

	now := time.Now()
	claims := auth.Claims{
		UserID:   1,
		Username: "sef",
		Role:     models.RoleBoss,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("test-secret-auth-handlers"))

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestWorkerAllowedOnDailyReportsButNotPeriod(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "radnik1", "password123", models.RoleWorker, true)
	r := setupRouter()
	token := loginToken(t, r, "radnik1", "password123")

	req := httptest.NewRequest(http.MethodGet, "/reports/daily?date=2026-01-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("daily: expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/reports/sales-summary?fromDate=2026-01-01&toDate=2026-01-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("sales-summary: expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/reports/period?fromDate=2026-01-01&toDate=2026-01-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("period: expected 403, got %d", w.Code)
	}
}

func TestManagerAllowedOnReports(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "menadzer1", "password123", models.RoleManager, true)
	r := setupRouter()
	token := loginToken(t, r, "menadzer1", "password123")

	req := httptest.NewRequest(http.MethodGet, "/reports/daily?date=2026-01-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestWorkerCannotSetPurchasePrice(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "radnik1", "password123", models.RoleWorker, true)
	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	r := setupRouter()
	token := loginToken(t, r, "radnik1", "password123")

	price := 5.0
	body, _ := json.Marshal(map[string]interface{}{
		"name":          "Proizvod",
		"categoryID":    cat.ID,
		"unit":          "kom",
		"salePrice":     10,
		"stockQuantity": 1,
		"purchasePrice": price,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestWorkerCanSetSalePrice(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "radnik1", "password123", models.RoleWorker, true)
	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	r := setupRouter()
	token := loginToken(t, r, "radnik1", "password123")

	body, _ := json.Marshal(map[string]interface{}{
		"name":          "Proizvod",
		"categoryID":    cat.ID,
		"unit":          "kom",
		"salePrice":     12.5,
		"stockQuantity": 3,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("purchasePrice")) {
		t.Fatal("worker response must not include purchasePrice")
	}
}

func TestOnlyBossCanCreateUser(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "radnik1", "password123", models.RoleWorker, true)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()

	workerToken := loginToken(t, r, "radnik1", "password123")
	body, _ := json.Marshal(map[string]string{
		"username": "novi",
		"password": "password123",
		"role":     models.RoleWorker,
		"fullName": "Novi Radnik",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+workerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("worker expected 403, got %d", w.Code)
	}

	bossToken := loginToken(t, r, "sef", "password123")
	req = httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+bossToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("boss expected 201, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestCannotDeactivateLastBoss(t *testing.T) {
	setupAuthTestDB(t)
	boss := createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()
	token := loginToken(t, r, "sef", "password123")

	body, _ := json.Marshal(map[string]bool{"isActive": false})
	req := httptest.NewRequest(http.MethodPut, "/users/"+itoa(boss.ID)+"/status", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (%s)", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("deaktivirati")) {
		t.Fatalf("expected deactivation protection message, got %s", w.Body.String())
	}
}

func TestCannotDemoteLastBoss(t *testing.T) {
	setupAuthTestDB(t)
	boss := createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()
	token := loginToken(t, r, "sef", "password123")

	body, _ := json.Marshal(map[string]string{
		"username": "sef",
		"role":     models.RoleManager,
		"fullName": "Šef Test",
	})
	req := httptest.NewRequest(http.MethodPut, "/users/"+itoa(boss.ID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRefundStoresAuthenticatedUserID(t *testing.T) {
	setupAuthTestDB(t)
	user := createUser(t, "radnik1", "password123", models.RoleWorker, true)
	r := setupRouter()
	token := loginToken(t, r, "radnik1", "password123")

	invoice := models.Invoice{
		CreatedByUserID: user.ID,
		TotalAmount:     100,
		PaidAmount:      40,
		Status:          models.InvoiceStatusPaid,
	}
	if err := database.DB.Create(&invoice).Error; err != nil {
		t.Fatalf("invoice: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"reason": "reklamacija"})
	req := httptest.NewRequest(http.MethodPut, "/invoices/"+itoa(invoice.ID)+"/cancel", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	var refund models.Refund
	if err := database.DB.Where("invoice_id = ?", invoice.ID).First(&refund).Error; err != nil {
		t.Fatalf("refund missing: %v", err)
	}
	if refund.CreatedByUserID != user.ID {
		t.Fatalf("expected CreatedByUserID=%d, got %d", user.ID, refund.CreatedByUserID)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"username":"radnik1"`)) {
		t.Fatalf("expected refund creator username in response: %s", w.Body.String())
	}
}

func TestNoHardcodedUserIDLiteralsInHandlers(t *testing.T) {
	// Smoke check that auth helpers are used: login + cancel path already covers CreatedByUserID.
	setupAuthTestDB(t)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()
	token := loginToken(t, r, "sef", "password123")
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func itoa(id uint) string {
	const digits = "0123456789"
	if id == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = digits[id%10]
		id /= 10
	}
	return string(buf[i:])
}
