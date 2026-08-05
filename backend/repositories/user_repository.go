package repositories

import (
	"errors"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"gorm.io/gorm"
)

func GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := database.DB.Order("id ASC").Find(&users).Error
	return users, err
}

func GetUserByID(id uint) (models.User, error) {
	var user models.User
	result := database.DB.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("korisnik nije pronađen")
		}
		return models.User{}, result.Error
	}
	return user, nil
}

func GetUserByUsername(username string) (models.User, error) {
	var user models.User
	result := database.DB.Where("username = ?", auth.NormalizeUsername(username)).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.User{}, errors.New("korisnik nije pronađen")
		}
		return models.User{}, result.Error
	}
	return user, nil
}

func CreateUser(user *models.User) error {
	user.Username = auth.NormalizeUsername(user.Username)
	if user.Username == "" {
		return errors.New("username je obavezan")
	}
	if !models.IsValidRole(user.Role) {
		return errors.New("nevalidna uloga")
	}

	var existing models.User
	err := database.DB.Where("username = ?", user.Username).First(&existing).Error
	if err == nil {
		return errors.New("username već postoji")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return database.DB.Select("Username", "PasswordHash", "Role", "FullName", "IsActive").Create(user).Error
}

func UpdateUser(user *models.User) error {
	user.Username = auth.NormalizeUsername(user.Username)
	if user.Username == "" {
		return errors.New("username je obavezan")
	}
	if !models.IsValidRole(user.Role) {
		return errors.New("nevalidna uloga")
	}

	var existing models.User
	err := database.DB.Where("username = ? AND id <> ?", user.Username, user.ID).First(&existing).Error
	if err == nil {
		return errors.New("username već postoji")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return database.DB.Model(user).
		Select("Username", "Role", "FullName", "IsActive", "PasswordHash").
		Updates(user).Error
}

func CountActiveBosses() (int64, error) {
	var count int64
	err := database.DB.Model(&models.User{}).
		Where("role = ? AND is_active = ?", models.RoleBoss, true).
		Count(&count).Error
	return count, err
}

func CountActiveBossesExcluding(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.User{}).
		Where("role = ? AND is_active = ? AND id <> ?", models.RoleBoss, true, userID).
		Count(&count).Error
	return count, err
}
