package main

import (
	"log"
	"os"


	"github.com/gin-gonic/gin"	
	"github.com/joho/godotenv"
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"am-keramika-backend/handlers"
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
		&models.Product{},
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

	r.POST("/products", handlers.CreateProduct)
	r.GET("/products", handlers.GetAllProducts)

	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("Neuspjela pokretanje servera: ", err)
	}
}
