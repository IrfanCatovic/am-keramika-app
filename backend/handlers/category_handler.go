package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/utils"

	"github.com/gin-gonic/gin"
)

func mapCategoryResponse(category models.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		IsActive:  category.IsActive,
		CreatedAt: category.CreatedAt.Format("2006-01-02 15:04"),
	}
}

func isCategoryValidationError(err error) bool {
	return errors.Is(err, repositories.ErrCategoryDuplicateName) ||
		errors.Is(err, repositories.ErrCategoryDuplicateSlug)
}

func isCategoryNotFoundError(err error) bool {
	return errors.Is(err, repositories.ErrCategoryNotFound)
}

func CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	category := &models.Category{
		Name: strings.TrimSpace(req.Name),
		Slug: slug,
	}

	if err := repositories.CreateCategory(category); err != nil {
		status := http.StatusInternalServerError
		if isCategoryValidationError(err) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": "Greška pri kreiranju kategorije", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kategorija kreirana",
		"data":    mapCategoryResponse(*category),
	})
}

func GetCategories(c *gin.Context) {
	includeInactive := c.Query("includeInactive") == "true"

	categories, err := repositories.GetCategories(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri dobavljanju kategorija", "error": err.Error()})
		return
	}

	response := make([]dto.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, mapCategoryResponse(category))
	}

	c.JSON(http.StatusOK, response)
}

func GetCategoryById(c *gin.Context) {
	category, err := repositories.GetCategoryByID(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if isCategoryNotFoundError(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Greška pri dobavljanju kategorije", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCategoryResponse(*category))
}

func UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID kategorije"})
		return
	}

	category, err := repositories.GetCategoryByID(strconv.FormatUint(id, 10))
	if err != nil {
		status := http.StatusInternalServerError
		if isCategoryNotFoundError(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Greška pri učitavanju kategorije", "error": err.Error()})
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	slug := utils.GenerateSlug(req.Name)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Slug nije validan"})
		return
	}

	category.Name = strings.TrimSpace(req.Name)
	category.Slug = slug

	if err := repositories.UpdateCategory(category); err != nil {
		status := http.StatusInternalServerError
		if isCategoryValidationError(err) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": "Greška pri ažuriranju kategorije", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategorija ažurirana",
		"data":    mapCategoryResponse(*category),
	})
}

func UpdateCategoryStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID kategorije"})
		return
	}

	var req dto.UpdateCategoryStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci", "error": err.Error()})
		return
	}

	if err := repositories.UpdateCategoryStatus(uint(id), req.IsActive); err != nil {
		status := http.StatusInternalServerError
		if isCategoryNotFoundError(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"message": "Greška pri promeni statusa kategorije", "error": err.Error()})
		return
	}

	category, err := repositories.GetCategoryByID(strconv.FormatUint(id, 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri učitavanju kategorije", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status kategorije ažuriran",
		"data":    mapCategoryResponse(*category),
	})
}

func DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID kategorije"})
		return
	}

	if err := repositories.DeleteCategory(uint(id)); err != nil {
		status := http.StatusInternalServerError
		switch {
		case isCategoryNotFoundError(err):
			status = http.StatusNotFound
		case errors.Is(err, repositories.ErrCategoryHasGroupsOrProducts):
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"message": "Greška pri brisanju kategorije", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kategorija obrisana"})
}
