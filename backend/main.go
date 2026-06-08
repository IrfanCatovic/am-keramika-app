package main

import (
	"log"
	"os"
	"net/http"

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

	r := gin.Default()
	r.Run(":" + port)
}
