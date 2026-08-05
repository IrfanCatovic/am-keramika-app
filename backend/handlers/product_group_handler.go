package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/utils"

	"github.com/gin-gonic/gin"
)

func CreateProductGroup(c *gin.Context) {
	var req dto.CreateProductGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(req.Name)
	}
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	group := models.ProductGroup{
		Name:       strings.TrimSpace(req.Name),
		Slug:       slug,
		CategoryID: req.CategoryID,
	}

	err := repositories.CreateProductGroup(&group)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "kategorija nije pronađena") ||
			strings.Contains(err.Error(), "već postoji") ||
			strings.Contains(err.Error(), "ima proizvode") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": "Greška pri kreiranju grupe", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Grupa proizvoda kreirana",
		"data": dto.ProductGroupResponse{
			ID:         group.ID,
			Name:       group.Name,
			Slug:       group.Slug,
			CategoryID: group.CategoryID,
		},
	})
}

func GetAllProductGroups(c *gin.Context) {
	categoryID := c.Query("categoryID")
	groups, err := repositories.GetAllProductGroups(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Greška pri učitavanju grupa",
			"error":   err.Error(),
		})
		return
	}

	response := make([]dto.ProductGroupListResponse, 0, len(groups))
	for _, group := range groups {
		response = append(response, dto.ProductGroupListResponse{
			ID:         group.ID,
			Name:       group.Name,
			Slug:       group.Slug,
			CategoryID: group.CategoryID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func GetProductGroupByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID grupe"})
		return
	}

	group, err := repositories.GetProductGroupByID(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "nije pronađena") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Greška pri učitavanju grupe", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ProductGroupResponse{
		ID:         group.ID,
		Name:       group.Name,
		Slug:       group.Slug,
		CategoryID: group.CategoryID,
	})
}

func UpdateProductGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID grupe"})
		return
	}

	group, err := repositories.GetProductGroupByID(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "nije pronađena") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Greška pri učitavanju grupe", "error": err.Error()})
		return
	}

	var req dto.UpdateProductGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(req.Name)
	}
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	group.Name = strings.TrimSpace(req.Name)
	group.Slug = slug
	group.CategoryID = req.CategoryID

	err = repositories.UpdateProductGroup(group)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case strings.Contains(err.Error(), "premjestite ili uklonite proizvode iz grupe prije promjene kategorije"):
			status = http.StatusConflict
		case strings.Contains(err.Error(), "kategorija nije pronađena"),
			strings.Contains(err.Error(), "već postoji"):
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": "Greška pri ažuriranju grupe", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grupa proizvoda ažurirana",
		"data": dto.ProductGroupResponse{
			ID:         group.ID,
			Name:       group.Name,
			Slug:       group.Slug,
			CategoryID: group.CategoryID,
		},
	})
}

func DeleteProductGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID grupe"})
		return
	}

	err = repositories.DeleteProductGroup(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "nije pronađena") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "ima proizvode") {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"message": "Greška pri brisanju grupe", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grupa proizvoda obrisana"})
}
