package handlers

import ("github.com/gin-gonic/gin"
	"net/http"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories")

func CreateCategory(c *gin.Context) {
	var category models.Category
	err := c.ShouldBindJSON(&category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message":"nesipavni podaci",}) 
		return
	}

	err = repositories.CreateCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message":"greska pri kreiranju kategorije",}) 
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message":"kategorija kreirana",}) 
	return
}

func GetCategories(c *gin.Context) {
	categories, err := repositories.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message":"greska pri dobavljanju kategorija",}) 
		return
	}
	c.JSON(http.StatusOK, categories)
	return
}

func GetCategoryById(c *gin.Context) {

	id := c.Param("id")
	category, err := repositories.GetCategoryByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message":"greska pri dobavljanju kategorije",}) 
		return
	}
	c.JSON(http.StatusOK, category)
	return
}
