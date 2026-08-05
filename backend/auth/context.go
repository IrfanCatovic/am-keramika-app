package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey   = "userID"
	ContextUsernameKey = "username"
	ContextRoleKey     = "role"
)

func SetAuthContext(c *gin.Context, userID uint, username, role string) {
	c.Set(ContextUserIDKey, userID)
	c.Set(ContextUsernameKey, username)
	c.Set(ContextRoleKey, role)
}

func GetUserID(c *gin.Context) (uint, error) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, errors.New("korisnik nije autentifikovan")
	}
	userID, ok := value.(uint)
	if !ok || userID == 0 {
		return 0, errors.New("nevalidan korisnički ID")
	}
	return userID, nil
}

func GetUsername(c *gin.Context) (string, error) {
	value, exists := c.Get(ContextUsernameKey)
	if !exists {
		return "", errors.New("korisnik nije autentifikovan")
	}
	username, ok := value.(string)
	if !ok || username == "" {
		return "", errors.New("nevalidan username")
	}
	return username, nil
}

func GetRole(c *gin.Context) (string, error) {
	value, exists := c.Get(ContextRoleKey)
	if !exists {
		return "", errors.New("korisnik nije autentifikovan")
	}
	role, ok := value.(string)
	if !ok || role == "" {
		return "", errors.New("nevalidna uloga")
	}
	return role, nil
}
