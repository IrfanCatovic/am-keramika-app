package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			sslmode,
		)
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Neuspela konekcija na bazu: ", err)
	}
	DB = database
	log.Println("Uspešna konekcija na bazu")
}

func Ping() error {
	if DB == nil {
		return fmt.Errorf("database nije inicijalizovan")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
