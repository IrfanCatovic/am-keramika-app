package auth

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const MinPasswordLength = 8

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func ValidatePassword(password string) error {
	if len(strings.TrimSpace(password)) < MinPasswordLength {
		return errors.New("lozinka mora imati najmanje 8 karaktera")
	}
	for _, r := range password {
		if unicode.IsSpace(r) && r != ' ' {
			return errors.New("lozinka sadrži nedozvoljene razmake")
		}
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("neuspelo heširanje lozinke")
	}
	return string(hashed), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
