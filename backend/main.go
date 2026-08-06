package main

import (
	"log"
	"os"

	"am-keramika-backend/auth"
	"am-keramika-backend/config"
	"am-keramika-backend/database"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"
	"am-keramika-backend/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	if err := config.RequireJWTSecret(); err != nil {
		log.Fatal(err)
	}

	corsOrigins, err := config.CORSAllowedOrigins()
	if err != nil {
		log.Fatal(err)
	}

	cloudName, apiKey, apiSecret, err := config.RequireCloudinary()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database.ConnectDB()
	err = database.DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.ProductGroup{},
		&models.Product{},
		&models.ProductImage{},
		&models.InventoryMovement{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Customer{},
		&models.Payment{},
		&models.PaymentAllocation{},
		&models.InvoiceCancellation{},
		&models.Refund{},
	)
	if err != nil {
		log.Fatal("Neuspjela migracija modela: ", err)
	}

	if err := auth.EnsureInitialBoss(); err != nil {
		log.Fatal(err)
	}

	imageStorage, err := storage.NewCloudinaryStorage(cloudName, apiKey, apiSecret)
	if err != nil {
		log.Fatal(err)
	}
	handlers.SetImageStorage(imageStorage)

	r := gin.Default()
	r.Use(middleware.CORS(corsOrigins))

	r.GET("/health", handlers.Health)
	r.POST("/auth/login", handlers.Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		authorized.GET("/auth/me", handlers.GetMe)

		bossOnly := authorized.Group("/")
		bossOnly.Use(middleware.RequireRoles(models.RoleBoss))
		{
			bossOnly.GET("/users", handlers.GetUsers)
			bossOnly.POST("/users", handlers.CreateUser)
			bossOnly.PUT("/users/:id", handlers.UpdateUser)
			bossOnly.PUT("/users/:id/password", handlers.UpdateUserPassword)
			bossOnly.PUT("/users/:id/status", handlers.UpdateUserStatus)
		}

		reports := authorized.Group("/reports")
		reports.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager))
		{
			reports.GET("/daily", handlers.GetDailyReport)
			reports.GET("/period", handlers.GetPeriodReport)
			reports.GET("/sales-summary", handlers.GetSalesSummaryReport)
			reports.GET("/transactions", handlers.GetFinancialTransactionsReport)
		}

		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/categories", handlers.CreateCategory)
			staff.GET("/categories", handlers.GetCategories)
			staff.GET("/categories/:id", handlers.GetCategoryById)

			staff.POST("/products", handlers.CreateProduct)
			staff.GET("/products", handlers.GetAllProducts)
			staff.GET("/products/:id", handlers.GetProductById)
			staff.PUT("/products/:id", handlers.UpdateProduct)
			staff.PUT("/products/:id/deactivate", handlers.DeactivateProduct)

			staff.POST("/products/:id/images", handlers.UploadProductImages)
			staff.PUT("/products/:productID/images/:imageID/primary", handlers.SetPrimaryProductImage)
			staff.PUT("/products/:productID/images/reorder", handlers.ReorderProductImages)
			staff.DELETE("/products/:productID/images/:imageID", handlers.DeleteProductImage)

			staff.POST("/product-groups", handlers.CreateProductGroup)
			staff.GET("/product-groups", handlers.GetAllProductGroups)
			staff.GET("/product-groups/:id", handlers.GetProductGroupByID)
			staff.PUT("/product-groups/:id", handlers.UpdateProductGroup)
			staff.DELETE("/product-groups/:id", handlers.DeleteProductGroup)

			staff.POST("/inventory/add", handlers.AddStock)
			staff.POST("/inventory/adjust", handlers.AdjustStock)
			staff.POST("/inventory/sell", handlers.SellStock)

			staff.POST("/invoices", handlers.CreateInvoice)
			staff.GET("/invoices/:id", handlers.GetInvoiceByID)
			staff.GET("/invoices", handlers.GetAllInvoices)
			staff.PUT("/invoices/:id/cancel", handlers.CancelInvoice)

			staff.POST("/customers", handlers.CreateCustomer)
			staff.GET("/customers", handlers.GetAllCustomers)
			staff.GET("/customers/:id", handlers.GetCustomerByID)
			staff.GET("/customers/:id/open-invoices", handlers.GetCustomerOpenInvoices)
			staff.GET("/customers/:id/payments", handlers.GetCustomerPayments)

			staff.POST("/payments", handlers.CreatePayment)
			staff.GET("/payments/:id", handlers.GetPaymentByID)
		}

		financeStaff := authorized.Group("/")
		financeStaff.Use(middleware.RequireRoles(models.RoleBoss, models.RoleManager))
		{
			financeStaff.GET("/customers/:id/financial-summary", handlers.GetCustomerFinancialSummary)
		}
	}

	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("Neuspjela pokretanje servera: ", err)
	}
}
