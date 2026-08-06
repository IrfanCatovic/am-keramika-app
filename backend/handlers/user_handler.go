package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/auth"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func mapUserResponse(user models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		FullName: user.FullName,
		IsActive: user.IsActive,
	}
}

func rejectBossTargetingDeveloper(c *gin.Context, target models.User) bool {
	if target.Role != models.RoleDeveloper {
		return false
	}
	actorRole, err := auth.GetRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return true
	}
	if actorRole == models.RoleBoss {
		c.JSON(http.StatusForbidden, gin.H{"message": "Šef ne smije mijenjati developer nalog"})
		return true
	}
	return false
}

func GetUsers(c *gin.Context) {
	users, err := repositories.GetManagedUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo dobavljanje korisnika"})
		return
	}

	response := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, mapUserResponse(user))
	}
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}

	if !models.IsAssignableUserRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Nevalidna uloga"})
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo kreiranje korisnika"})
		return
	}

	user := models.User{
		Username:     req.Username,
		PasswordHash: hash,
		Role:         req.Role,
		FullName:     strings.TrimSpace(req.FullName),
		IsActive:     true,
	}

	err = repositories.CreateUser(&user)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "već postoji") || strings.Contains(err.Error(), "nevalidna") {
			if strings.Contains(err.Error(), "već postoji") {
				status = http.StatusConflict
			} else {
				status = http.StatusBadRequest
			}
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Korisnik kreiran",
		"data":    mapUserResponse(user),
	})
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID korisnika"})
		return
	}

	user, err := repositories.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Korisnik nije pronađen"})
		return
	}
	if rejectBossTargetingDeveloper(c, user) {
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}
	if !models.IsAssignableUserRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Nevalidna uloga"})
		return
	}
	if user.Role == models.RoleDeveloper {
		c.JSON(http.StatusForbidden, gin.H{"message": "Developer nalog se ne smije mijenjati kroz ovaj endpoint"})
		return
	}

	if user.Role == models.RoleBoss && user.IsActive && req.Role != models.RoleBoss {
		remaining, countErr := repositories.CountActiveBossesExcluding(user.ID)
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo ažuriranje korisnika"})
			return
		}
		if remaining == 0 {
			c.JSON(http.StatusConflict, gin.H{"message": "Nije dozvoljeno degradirati posljednjeg aktivnog šefa"})
			return
		}
	}

	user.Username = req.Username
	user.Role = req.Role
	user.FullName = strings.TrimSpace(req.FullName)

	err = repositories.UpdateUser(&user)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "već postoji") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "nevalidna") || strings.Contains(err.Error(), "obavezan") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Korisnik ažuriran",
		"data":    mapUserResponse(user),
	})
}

func UpdateUserPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID korisnika"})
		return
	}

	user, err := repositories.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Korisnik nije pronađen"})
		return
	}
	if rejectBossTargetingDeveloper(c, user) {
		return
	}
	if user.Role == models.RoleDeveloper {
		actorRole, roleErr := auth.GetRole(c)
		if roleErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
			return
		}
		if actorRole != models.RoleDeveloper {
			c.JSON(http.StatusForbidden, gin.H{"message": "Developer nalog se ne smije mijenjati kroz ovaj endpoint"})
			return
		}
	}

	var req dto.UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjela promjena lozinke"})
		return
	}

	user.PasswordHash = hash
	if err := repositories.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjela promjena lozinke"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lozinka je ažurirana"})
}

func UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID korisnika"})
		return
	}

	currentUserID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
		return
	}

	user, err := repositories.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Korisnik nije pronađen"})
		return
	}
	if rejectBossTargetingDeveloper(c, user) {
		return
	}
	if user.Role == models.RoleDeveloper {
		c.JSON(http.StatusForbidden, gin.H{"message": "Developer nalog se ne smije mijenjati kroz ovaj endpoint"})
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}

	if !req.IsActive && user.ID == currentUserID {
		c.JSON(http.StatusConflict, gin.H{"message": "Ne možete deaktivirati vlastiti nalog"})
		return
	}

	if user.Role == models.RoleBoss && user.IsActive && !req.IsActive {
		remaining, countErr := repositories.CountActiveBossesExcluding(user.ID)
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjela izmjena statusa"})
			return
		}
		if remaining == 0 {
			c.JSON(http.StatusConflict, gin.H{"message": "Nije dozvoljeno deaktivirati posljednjeg aktivnog šefa"})
			return
		}
	}

	user.IsActive = req.IsActive
	if err := repositories.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjela izmjena statusa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status korisnika ažuriran",
		"data":    mapUserResponse(user),
	})
}
