package auth

import (
	"errors"
	"log"
	"os"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

// EnsureInitialBoss kreira početnog šefa ako nijedan ne postoji.
// Idempotentno: ne kreira novog šefa ako već postoji bar jedan sa ulogom sef.
func EnsureInitialBoss() error {
	var count int64
	if err := database.DB.Model(&models.User{}).Where("role = ?", models.RoleBoss).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	username := NormalizeUsername(os.Getenv("INITIAL_BOSS_USERNAME"))
	password := os.Getenv("INITIAL_BOSS_PASSWORD")
	if username == "" || password == "" {
		return errors.New("nema šefa u sistemu; postavite INITIAL_BOSS_USERNAME i INITIAL_BOSS_PASSWORD")
	}
	if err := ValidatePassword(password); err != nil {
		return errors.New("INITIAL_BOSS_PASSWORD: " + err.Error())
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	boss := models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         models.RoleBoss,
		IsActive:     true,
		FullName:     "Šef",
	}

	if err := database.DB.Create(&boss).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}

	log.Printf("Kreiran početni šef korisnik: %s", username)
	return nil
}
