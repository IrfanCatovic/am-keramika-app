package database


import (
	"fmt"
	"log"
	"os"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"	
	"github.com/joho/godotenv"
	"am-keramika-backend/models"
)

var DB *gorm.DB //Global database connection

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Nije pronađen .env file")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
	os.Getenv("DB_HOST"), 
	os.Getenv("DB_PORT"), 
	os.Getenv("DB_USER"), 
	os.Getenv("DB_PASSWORD"), 
	os.Getenv("DB_NAME"),
)

database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
	log.Fatal("Neuspjela konekcija na bazu: ", err)
}
	DB = database
	fmt.Println("Uspješna konekcija na bazu")
}

