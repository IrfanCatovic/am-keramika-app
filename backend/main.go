package main

import (
	"log"
	"os"

	"am-keramika-backend/database"
	"am-keramika-backend/handlers"
	"am-keramika-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Nije pronađen .env file")
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

	r := gin.Default()
	r.GET("/users", handlers.GetUsers)
	r.POST("/users", handlers.CreateUser)
	r.GET("/users/:username", handlers.GetUserByUsername)
	r.DELETE("/users/:id", handlers.DeleteUser)

	r.POST("/categories", handlers.CreateCategory)
	r.GET("/categories", handlers.GetCategories)
	r.GET("/categories/:id", handlers.GetCategoryById)
	r.GET("/customers/:id/open-invoices", handlers.GetCustomerOpenInvoices)

	r.POST("/products", handlers.CreateProduct)
	r.GET("/products", handlers.GetAllProducts)
	r.GET("/products/:id", handlers.GetProductById)
	r.PUT("/products/:id", handlers.UpdateProduct)
	r.PUT("/products/:id/deactivate", handlers.DeactivateProduct)

	r.POST("/product-groups", handlers.CreateProductGroup)
	r.GET("/product-groups", handlers.GetAllProductGroups)
	r.GET("/product-groups/:id", handlers.GetProductGroupByID)
	r.PUT("/product-groups/:id", handlers.UpdateProductGroup)
	r.DELETE("/product-groups/:id", handlers.DeleteProductGroup)

	r.POST("/inventory/add", handlers.AddStock)
	r.POST("/inventory/adjust", handlers.AdjustStock)
	r.POST("/inventory/sell", handlers.SellStock)

	r.POST("/invoices", handlers.CreateInvoice)
	r.GET("/invoices/:id", handlers.GetInvoiceByID)
	r.GET("/invoices", handlers.GetAllInvoices)
	r.PUT("/invoices/:id/cancel", handlers.CancelInvoice)

	r.POST("/customers", handlers.CreateCustomer)
	r.GET("/customers", handlers.GetAllCustomers)
	r.GET("/customers/:id", handlers.GetCustomerByID)
	r.GET("/customers/:id/financial-summary", handlers.GetCustomerFinancialSummary)
	r.GET("/customers/:id/payments", handlers.GetCustomerPayments)

	r.POST("/payments", handlers.CreatePayment)
	r.GET("/payments/:id", handlers.GetPaymentByID)

	r.GET("/reports/daily", handlers.GetDailyReport)
	r.GET("/reports/period", handlers.GetPeriodReport)
	r.GET("/reports/sales-summary", handlers.GetSalesSummaryReport)
	r.GET("/reports/transactions", handlers.GetFinancialTransactionsReport)

	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("Neuspjela pokretanje servera: ", err)
	}
}
