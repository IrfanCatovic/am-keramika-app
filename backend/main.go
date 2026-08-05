package main

import (
	"log"
	"os"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/handlers"
	"am-keramika-backend/middleware"
	"am-keramika-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Nije pronađen .env file")
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET nije postavljen")
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

	r := gin.Default()

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
