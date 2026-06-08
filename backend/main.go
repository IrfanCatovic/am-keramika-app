package main

import (
	"log"
	"os"


	"github.com/gin-gonic/gin"	
	"github.com/joho/godotenv"
	"am-keramika-backend/database"
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
	r := gin.Default()
	r.Run(":" + port)
}
