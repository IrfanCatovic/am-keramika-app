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

	if err := auth.EnsureInitialDeveloper(); err != nil {
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

		userAdmin := authorized.Group("/")
		userAdmin.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss))
		{
			userAdmin.GET("/users", handlers.GetUsers)
			userAdmin.POST("/users", handlers.CreateUser)
			userAdmin.PUT("/users/:id", handlers.UpdateUser)
			userAdmin.PUT("/users/:id/password", handlers.UpdateUserPassword)
			userAdmin.PUT("/users/:id/status", handlers.UpdateUserStatus)
		}

		reports := authorized.Group("/reports")
		reports.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager))
		{
			reports.GET("/daily", handlers.GetDailyReport)
			reports.GET("/period", handlers.GetPeriodReport)
			reports.GET("/sales-summary", handlers.GetSalesSummaryReport)
			reports.GET("/transactions", handlers.GetFinancialTransactionsReport)
		}

		staff := authorized.Group("/")
		staff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager, models.RoleWorker))
		{
			staff.POST("/categories", handlers.CreateCategory)
			staff.GET("/categories", handlers.GetCategories)
			staff.GET("/categories/:id", handlers.GetCategoryById)
			staff.PUT("/categories/:id", handlers.UpdateCategory)
			staff.PUT("/categories/:id/status", handlers.UpdateCategoryStatus)
			staff.DELETE("/categories/:id", handlers.DeleteCategory)

			staff.POST("/products", handlers.CreateProduct)
			staff.GET("/products", handlers.GetAllProducts)
			staff.GET("/products/:id", handlers.GetProductById)
			staff.PUT("/products/:id", handlers.UpdateProduct)
			staff.PUT("/products/:id/deactivate", handlers.DeactivateProduct)
			staff.PUT("/products/:id/activate", handlers.ActivateProduct)

			staff.POST("/products/:id/images", handlers.UploadProductImages)
			staff.PUT("/products/:id/images/:imageID/primary", handlers.SetPrimaryProductImage)
			staff.PUT("/products/:id/images/reorder", handlers.ReorderProductImages)
			staff.DELETE("/products/:id/images/:imageID", handlers.DeleteProductImage)

			staff.POST("/product-groups", handlers.CreateProductGroup)
			staff.GET("/product-groups", handlers.GetAllProductGroups)
			staff.GET("/product-groups/:id", handlers.GetProductGroupByID)
			staff.PUT("/product-groups/:id", handlers.UpdateProductGroup)
			staff.DELETE("/product-groups/:id", handlers.DeleteProductGroup)

			staff.POST("/inventory/add", handlers.AddStock)
			staff.POST("/inventory/adjust", handlers.AdjustStock)
			staff.POST("/inventory/sell", handlers.SellStock)
			staff.GET("/inventory/low-stock", handlers.GetLowStock)
			staff.GET("/inventory/summary", handlers.GetInventorySummary)
			staff.GET("/inventory/movements", handlers.GetInventoryMovements)

			staff.POST("/invoices", handlers.CreateInvoice)
			staff.GET("/invoices/:id/pdf", handlers.GetInvoicePDF)
			staff.GET("/invoices/:id", handlers.GetInvoiceByID)
			staff.GET("/invoices", handlers.GetAllInvoices)
			staff.PUT("/invoices/:id/cancel", handlers.CancelInvoice)

			staff.POST("/customers", handlers.CreateCustomer)
			staff.GET("/customers", handlers.GetAllCustomers)
			staff.GET("/customers/:id", handlers.GetCustomerByID)
			staff.PUT("/customers/:id", handlers.UpdateCustomer)
			staff.PUT("/customers/:id/status", handlers.UpdateCustomerStatus)
			staff.DELETE("/customers/:id", handlers.DeleteCustomer)
			staff.GET("/customers/:id/open-invoices", handlers.GetCustomerOpenInvoices)
			staff.GET("/customers/:id/payments", handlers.GetCustomerPayments)

			staff.POST("/payments", handlers.CreatePayment)
			staff.GET("/payments", handlers.GetAllPayments)
			staff.GET("/payments/:id", handlers.GetPaymentByID)
		}

		financeStaff := authorized.Group("/")
		financeStaff.Use(middleware.RequireRoles(models.RoleDeveloper, models.RoleBoss, models.RoleManager))
		{
			financeStaff.GET("/customers/:id/financial-summary", handlers.GetCustomerFinancialSummary)
			financeStaff.GET("/refunds", handlers.GetRefunds)
		}
	}

	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("Neuspjela pokretanje servera: ", err)
	}
}
