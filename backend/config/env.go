package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv učitava backend/.env jednom. Ako fajl ne postoji, koriste se sistemske varijable.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Napomena: .env fajl nije učitan, koriste se sistemske environment varijable")
	}
}

func RequireJWTSecret() error {
	if os.Getenv("JWT_SECRET") == "" {
		return errors.New("JWT_SECRET nije postavljen")
	}
	return nil
}

func RequireCloudinary() (cloudName, apiKey, apiSecret string, err error) {
	cloudName = os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey = os.Getenv("CLOUDINARY_API_KEY")
	apiSecret = os.Getenv("CLOUDINARY_API_SECRET")

	switch {
	case cloudName == "":
		return "", "", "", errors.New("CLOUDINARY_CLOUD_NAME nije postavljen")
	case apiKey == "":
		return "", "", "", errors.New("CLOUDINARY_API_KEY nije postavljen")
	case apiSecret == "":
		return "", "", "", errors.New("CLOUDINARY_API_SECRET nije postavljen")
	}
	return cloudName, apiKey, apiSecret, nil
}
