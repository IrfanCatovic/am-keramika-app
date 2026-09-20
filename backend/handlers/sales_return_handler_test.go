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
		&models.ProductImage{},
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
			staff.GET("/sales-returns", handlers.GetSalesReturns)
			staff.GET("/sales-returns/:id", handlers.GetSalesReturnByID)
			staff.GET("/products", handlers.GetAllProducts)
		}
		finance := authorized.Group("/")
		finance.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager))
		{
			finance.GET("/refunds", handlers.GetRefunds)
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

func TestGetSalesReturnsWorkerCanListAndOpenDetail(t *testing.T) {
	setupSalesReturnHandlerTestDB(t)
	r := setupSalesReturnHandlerRouter()
	token, user := salesReturnLogin(t, r, "radnik1", models.RoleWorker)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	product := models.Product{
		Name: "Calacatta", Slug: "calacatta", CategoryID: cat.ID, Unit: "m²",
		SalePrice: 2000, StockQuantity: 30, IsActive: true,
		SaleByPackage: true, PackageQuantity: 1.44,
	}
	database.DB.Create(&product)

	older := httptest.NewRecorder()
	r.ServeHTTP(older, salesReturnJSONRequest(t, token, map[string]any{
		"description":  "stariji",
		"cashRefunded": false,
		"items":        []map[string]any{{"productID": product.ID, "quantity": 1, "unitPrice": 1600}},
	}))
	if older.Code != http.StatusCreated {
		t.Fatalf("older create=%d %s", older.Code, older.Body.String())
	}

	newer := httptest.NewRecorder()
	r.ServeHTTP(newer, salesReturnJSONRequest(t, token, map[string]any{
		"description":  "Marko Marković - višak",
		"cashRefunded": true,
		"items":        []map[string]any{{"productID": product.ID, "quantity": 20, "unitPrice": 1600}},
	}))
	if newer.Code != http.StatusCreated {
		t.Fatalf("newer create=%d %s", newer.Code, newer.Body.String())
	}
	var created dto.SalesReturnResponse
	json.Unmarshal(newer.Body.Bytes(), &created)

	listReq := httptest.NewRequest(http.MethodGet, "/sales-returns?page=1&pageSize=20", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("list=%d %s", listW.Code, listW.Body.String())
	}
	var list dto.PaginatedSalesReturnResponse
	json.Unmarshal(listW.Body.Bytes(), &list)
	if list.Page != 1 || list.PageSize != 20 || list.Total != 2 || len(list.Items) != 2 {
		t.Fatalf("list=%+v", list)
	}
	if list.Items[0].ID != created.ID || list.Items[0].CreatedByUser == nil || list.Items[0].CreatedByUser.ID != user.ID {
		t.Fatalf("newest/audit=%+v", list.Items[0])
	}

	cashReq := httptest.NewRequest(http.MethodGet, "/sales-returns?cashRefunded=true", nil)
	cashReq.Header.Set("Authorization", "Bearer "+token)
	cashW := httptest.NewRecorder()
	r.ServeHTTP(cashW, cashReq)
	var cashList dto.PaginatedSalesReturnResponse
	json.Unmarshal(cashW.Body.Bytes(), &cashList)
	if cashList.Total != 1 || !cashList.Items[0].CashRefunded {
		t.Fatalf("cash filter=%+v", cashList)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/sales-returns/"+strconv.FormatUint(uint64(created.ID), 10), nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailW := httptest.NewRecorder()
	r.ServeHTTP(detailW, detailReq)
	if detailW.Code != http.StatusOK {
		t.Fatalf("detail=%d %s", detailW.Code, detailW.Body.String())
	}
	var detail dto.SalesReturnResponse
	json.Unmarshal(detailW.Body.Bytes(), &detail)
	if detail.RefundID == nil || len(detail.Items) != 1 {
		t.Fatalf("detail=%+v", detail)
	}
	item := detail.Items[0]
	if item.ProductName != "Calacatta" || item.Quantity != 20 || !item.SaleByPackage || item.PackageQuantity != 1.44 {
		t.Fatalf("snapshot=%+v", item)
	}

	refundReq := httptest.NewRequest(http.MethodGet, "/refunds", nil)
	refundReq.Header.Set("Authorization", "Bearer "+token)
	refundW := httptest.NewRecorder()
	r.ServeHTTP(refundW, refundReq)
	if refundW.Code != http.StatusForbidden {
		t.Fatalf("worker /refunds want 403 got %d", refundW.Code)
	}
}

func TestGetAllProductsIncludeInactiveForReturnSearch(t *testing.T) {
	setupSalesReturnHandlerTestDB(t)
	r := setupSalesReturnHandlerRouter()
	token, _ := salesReturnLogin(t, r, "radnik1", models.RoleWorker)

	cat := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	database.DB.Create(&cat)
	inactive := models.Product{
		Name: "Stari bojler", Slug: "stari-bojler", CategoryID: cat.ID, Unit: "kom",
		SalePrice: 4500, StockQuantity: 2, IsActive: true,
	}
	database.DB.Create(&inactive)
	database.DB.Model(&inactive).Update("is_active", false)

	packageProduct := models.Product{
		Name: "Calacatta", Slug: "calacatta", CategoryID: cat.ID, Unit: "m²",
		SalePrice: 2000, StockQuantity: 10, IsActive: true,
		SaleByPackage: true, PackageQuantity: 1.44,
	}
	database.DB.Create(&packageProduct)

	hidden := httptest.NewRequest(http.MethodGet, "/products?search=stari", nil)
	hidden.Header.Set("Authorization", "Bearer "+token)
	hiddenW := httptest.NewRecorder()
	r.ServeHTTP(hiddenW, hidden)
	var hiddenResp dto.PaginatedProductListResponse
	json.Unmarshal(hiddenW.Body.Bytes(), &hiddenResp)
	if hiddenResp.Pagination.TotalItems != 0 {
		t.Fatalf("default search must hide inactive, got %+v", hiddenResp.Products)
	}

	req := httptest.NewRequest(http.MethodGet, "/products?search=stari&includeInactive=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("includeInactive=%d %s", w.Code, w.Body.String())
	}
	var resp dto.PaginatedProductListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Products) != 1 || resp.Products[0].IsActive || resp.Products[0].Unit != "kom" {
		t.Fatalf("inactive search=%+v", resp.Products)
	}
	if resp.Products[0].EffectiveSalePrice != 4500 {
		t.Fatalf("effectiveSalePrice=%v", resp.Products[0].EffectiveSalePrice)
	}

	pkgReq := httptest.NewRequest(http.MethodGet, "/products?search=calacatta", nil)
	pkgReq.Header.Set("Authorization", "Bearer "+token)
	pkgW := httptest.NewRecorder()
	r.ServeHTTP(pkgW, pkgReq)
	var pkg dto.PaginatedProductListResponse
	json.Unmarshal(pkgW.Body.Bytes(), &pkg)
	if len(pkg.Products) != 1 || !pkg.Products[0].SaleByPackage || pkg.Products[0].PackageQuantity != 1.44 {
		t.Fatalf("package fields=%+v", pkg.Products)
	}
	if pkg.Products[0].Unit != "m²" || pkg.Products[0].EffectiveSalePrice != 2000 {
		t.Fatalf("unit/price=%+v", pkg.Products[0])
	}
}

func salesReturnJSONRequest(t *testing.T, token string, payload map[string]any) *http.Request {
	t.Helper()
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/sales-returns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}
