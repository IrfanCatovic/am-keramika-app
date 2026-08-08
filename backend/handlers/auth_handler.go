package handlers

import (
	"net/http"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}

	username := auth.NormalizeUsername(req.Username)
	user, err := repositories.GetUserByUsername(username)
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Pogrešan username ili lozinka"})
		return
	}
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnički nalog je deaktiviran"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspelo generisanje tokena"})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token: token,
		User: dto.AuthUserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			FullName: user.FullName,
		},
	})
}

func ChangePassword(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}

	user, err := repositories.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije pronađen"})
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Trenutna lozinka nije tačna."})
		return
	}

	if req.NewPassword == req.CurrentPassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Nova lozinka mora biti drugačija od trenutne."})
		return
	}

	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Nova lozinka mora imati najmanje 8 karaktera."})
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Lozinku trenutno nije moguće promijeniti."})
		return
	}

	user.PasswordHash = hash
	user.TokenVersion++
	if err := repositories.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Lozinku trenutno nije moguće promijeniti."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lozinka je uspješno promijenjena. Prijavite se ponovo.",
	})
}

func GetMe(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	user, err := repositories.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije pronađen"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		FullName: user.FullName,
		IsActive: user.IsActive,
	})
}
