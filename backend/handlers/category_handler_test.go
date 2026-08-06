package handlers

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
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCategoryHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-category-handlers")
}

func setupCategoryTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/categories", CreateCategory)
			staff.GET("/categories", GetCategories)
			staff.GET("/categories/:id", GetCategoryById)
			staff.PUT("/categories/:id", UpdateCategory)
			staff.PUT("/categories/:id/status", UpdateCategoryStatus)
			staff.DELETE("/categories/:id", DeleteCategory)
		}
	}
	return r
}

func categoryLoginToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	user := models.User{Username: "radnik1", PasswordHash: hash, Role: models.RoleWorker, IsActive: true}
	if err := repositories.CreateUser(&user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"username": "radnik1", "password": "password123"})
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

func TestCreateCategoryReturnsDTO(t *testing.T) {
	setupCategoryHandlerTestDB(t)
	r := setupCategoryTestRouter()
	token := categoryLoginToken(t, r)

	body, _ := json.Marshal(dto.CreateCategoryRequest{Name: "Keramika"})
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}

	var resp struct {
		Data dto.CategoryResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.ID == 0 || resp.Data.Name != "Keramika" || resp.Data.Slug != "keramika" {
		t.Fatalf("unexpected DTO: %+v", resp.Data)
	}
	if !resp.Data.IsActive || resp.Data.CreatedAt == "" {
		t.Fatalf("expected active category with createdAt, got %+v", resp.Data)
	}
}
