package handlers

import (
	"am-keramika-backend/repositories"
	"am-keramika-backend/models"

	"net/http"
	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	users, err := repositories.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func CreateUser(c *gin.Context) {
 	var user models.User
	err := c.ShouldBindJSON(&user) //ovde citamo JSON iz requesta i smjestamo u user strukturu
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspjelo bindovanje JSON-a"})
		return
	}
	err = repositories.CreateUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo kreiranje korisnika"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Korisnik kreiran", "data": user})
}

func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Neuspjelo dobavljanje korisnika"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}