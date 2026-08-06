package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/storage"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var jpegBytes = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

func setupProductImageHandlerTest(t *testing.T) (*gin.Engine, models.Product, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Product{}, &models.ProductImage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db

	hash, _ := auth.HashPassword("password123")
	user := models.User{Username: "radnik1", PasswordHash: hash, Role: models.RoleWorker, IsActive: true}
	db.Create(&user)

	category := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	db.Create(&category)
	product := models.Product{Name: "Pločica", Slug: "plocica", CategoryID: category.ID, Unit: "kom", SalePrice: 10, IsActive: true}
	db.Create(&product)

	handlers.SetImageStorage(storage.NewFakeStorage())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handlers.Login)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired(), middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
	{
		authorized.POST("/products/:id/images", handlers.UploadProductImages)
		authorized.GET("/products", handlers.GetAllProducts)
		authorized.GET("/products/:id", handlers.GetProductById)
	}

	body, _ := json.Marshal(map[string]string{"username": "radnik1", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var login dto.LoginResponse
	json.Unmarshal(w.Body.Bytes(), &login)

	return r, product, login.Token
}

func TestProductDetailReturnsAllImages(t *testing.T) {
	r, product, token := setupProductImageHandlerTest(t)

	fake := storage.NewFakeStorage()
	handlers.SetImageStorage(fake)
	_, _ = repositories.UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{
		{Reader: bytes.NewReader(jpegBytes), MIMEType: "image/jpeg", Size: int64(len(jpegBytes))},
		{Reader: bytes.NewReader(jpegBytes), MIMEType: "image/jpeg", Size: int64(len(jpegBytes))},
	})

	req := httptest.NewRequest(http.MethodGet, "/products/"+strconv.FormatUint(uint64(product.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Images) != 2 {
		t.Fatalf("expected 2 images on detail, got %d", len(resp.Images))
	}
}

func TestProductListReturnsOnlyPrimary(t *testing.T) {
	r, product, token := setupProductImageHandlerTest(t)

	fake := storage.NewFakeStorage()
	handlers.SetImageStorage(fake)
	_, _ = repositories.UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{
		{Reader: bytes.NewReader(jpegBytes), MIMEType: "image/jpeg", Size: int64(len(jpegBytes))},
		{Reader: bytes.NewReader(jpegBytes), MIMEType: "image/jpeg", Size: int64(len(jpegBytes))},
	})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp []dto.ProductResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 product")
	}
	if resp[0].PrimaryImage == nil {
		t.Fatal("expected primaryImage on list")
	}
	if len(resp[0].Images) != 0 {
		t.Fatalf("list must not include images array, got %d", len(resp[0].Images))
	}
}

func TestMultipartUploadEndpoint(t *testing.T) {
	r, product, token := setupProductImageHandlerTest(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("images", "test.jpg")
	part.Write(jpegBytes)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/products/"+strconv.FormatUint(uint64(product.ID), 10)+"/images", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
}
