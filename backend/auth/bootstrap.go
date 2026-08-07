package auth

import (
	"errors"
	"log"
	"os"
	"strings"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

const defaultDeveloperFullName = "Irfan Catovic"

// EnsureInitialDeveloper kreira početnog developer nalog ako nijedan ne postoji.
// Idempotentno: ne kreira niti mijenja postojeći developer (uključujući soft-deleted),
// osim što dopunjava placeholder FullName za prikaz u auditu.
func EnsureInitialDeveloper() error {
	exists, err := developerExists()
	if err != nil {
		return err
	}
	if exists {
		return syncDeveloperDisplayName()
	}

	if !isDeveloperBootstrapEnabled() {
		return errors.New("nema developer naloga u sistemu; uključite ENABLE_DEVELOPER_BOOTSTRAP=true i postavite INITIAL_DEVELOPER_USERNAME i INITIAL_DEVELOPER_PASSWORD")
	}

	username := NormalizeUsername(os.Getenv("INITIAL_DEVELOPER_USERNAME"))
	password := os.Getenv("INITIAL_DEVELOPER_PASSWORD")
	if username == "" || password == "" {
		return errors.New("ENABLE_DEVELOPER_BOOTSTRAP je uključen, ali INITIAL_DEVELOPER_USERNAME i INITIAL_DEVELOPER_PASSWORD moraju biti postavljeni")
	}
	if err := ValidatePassword(password); err != nil {
		return errors.New("INITIAL_DEVELOPER_PASSWORD: " + err.Error())
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	developer := models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         models.RoleDeveloper,
		IsActive:     true,
		FullName:     developerFullName(),
	}

	if err := database.DB.Create(&developer).Error; err != nil {
		// Race / već postoji username: ponovo provjeri developer red (idempotentnost).
		if existsNow, checkErr := developerExists(); checkErr != nil {
			return checkErr
		} else if existsNow {
			return syncDeveloperDisplayName()
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("ne može se kreirati developer: username već postoji bez role developer")
		}
		return err
	}

	log.Printf("Kreiran početni developer korisnik: %s", username)
	return nil
}

func developerFullName() string {
	name := strings.TrimSpace(os.Getenv("INITIAL_DEVELOPER_FULL_NAME"))
	if name != "" {
		return name
	}
	return defaultDeveloperFullName
}

// syncDeveloperDisplayName sets a real display name when the bootstrap placeholder remains.
func syncDeveloperDisplayName() error {
	desired := developerFullName()
	return database.DB.Model(&models.User{}).
		Where("role = ?", models.RoleDeveloper).
		Where("full_name = ? OR full_name = ? OR full_name IS NULL", "Developer", "").
		Update("full_name", desired).Error
}

func isDeveloperBootstrapEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ENABLE_DEVELOPER_BOOTSTRAP")), "true")
}

func developerExists() (bool, error) {
	var count int64
	// Unscoped: uključi soft-deleted developer naloge da se ne kreira duplikat.
	err := database.DB.Unscoped().Model(&models.User{}).
		Where("role = ?", models.RoleDeveloper).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
