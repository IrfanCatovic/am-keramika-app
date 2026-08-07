package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
	"am-keramika-backend/storage"

	"github.com/gin-gonic/gin"
)

func mapProductImageResponse(image models.ProductImage) dto.ProductImageResponse {
	return dto.ProductImageResponse{
		ID:        image.ID,
		URL:       image.URL,
		IsPrimary: image.IsPrimary,
		SortOrder: image.SortOrder,
		Width:     image.Width,
		Height:    image.Height,
		Format:    image.Format,
	}
}

func UploadProductImages(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan ID proizvoda"})
		return
	}

	store := getImageStorage()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Storage nije konfigurisan"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan multipart zahtjev"})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Polje images je obavezno"})
		return
	}

	if err := repositories.ProductExists(uint(productID)); err != nil {
		if errors.Is(err, repositories.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Proizvod nije pronađen"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri provjeri proizvoda"})
		return
	}

	existingCount, err := repositories.CountProductImages(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Greška pri provjeri slika"})
		return
	}
	if existingCount+int64(len(files)) > storage.MaxImagesPerProduct {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Proizvod može imati najviše 8 slika"})
		return
	}

	validated := make([]storage.ValidatedImage, 0, len(files))
	for _, fileHeader := range files {
		if fileHeader.Size > storage.MaxImageSizeBytes {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Jedna ili više slika prelazi 10 MB"})
			return
		}
		file, openErr := fileHeader.Open()
		if openErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspelo otvaranje fajla"})
			return
		}
		valid, validateErr := storage.ValidateImageFile(fileHeader.Size, file)
		file.Close()
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": validateErr.Error()})
			return
		}
		validated = append(validated, valid)
	}

	uploadReaders := make([]storage.ValidatedImage, 0, len(files))
	for i, fileHeader := range files {
		file, openErr := fileHeader.Open()
		if openErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Neuspelo otvaranje fajla"})
			return
		}
		uploadFile, validateErr := storage.ValidateImageFile(fileHeader.Size, file)
		file.Close()
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": validateErr.Error()})
			return
		}
		uploadFile.MIMEType = validated[i].MIMEType
		uploadReaders = append(uploadReaders, uploadFile)
	}

	images, err := repositories.UploadProductImages(context.Background(), store, uint(productID), uploadReaders)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, repositories.ErrProductNotFound):
			status = http.StatusNotFound
		case errors.Is(err, repositories.ErrMaxImagesReached):
			status = http.StatusBadRequest
		case strings.Contains(err.Error(), "MIME"):
			status = http.StatusBadRequest
		case strings.Contains(err.Error(), "cloudinary"):
			status = http.StatusBadGateway
		}
		c.JSON(status, gin.H{"message": "Greška pri uploadu slika", "error": err.Error()})
		return
	}

	response := make([]dto.ProductImageResponse, 0, len(images))
	for _, image := range images {
		response = append(response, mapProductImageResponse(image))
	}
	c.JSON(http.StatusCreated, gin.H{"data": response})
}

func SetPrimaryProductImage(c *gin.Context) {
	productID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	imageID, err := parseUintParam(c, "imageID")
	if err != nil {
		return
	}

	image, err := repositories.SetPrimaryProductImage(productID, imageID)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, repositories.ErrImageNotFound):
			status = http.StatusNotFound
		case errors.Is(err, repositories.ErrImageWrongProduct):
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapProductImageResponse(*image)})
}

func ReorderProductImages(c *gin.Context) {
	productID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	var req dto.ReorderProductImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravni podaci"})
		return
	}

	images, err := repositories.ReorderProductImages(productID, req.ImageIDs)
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, repositories.ErrInvalidReorderRequest) {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	response := make([]dto.ProductImageResponse, 0, len(images))
	for _, image := range images {
		response = append(response, mapProductImageResponse(image))
	}
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func DeleteProductImage(c *gin.Context) {
	productID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	imageID, err := parseUintParam(c, "imageID")
	if err != nil {
		return
	}

	store := getImageStorage()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Storage nije konfigurisan"})
		return
	}

	err = repositories.DeleteProductImage(context.Background(), store, productID, imageID)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, repositories.ErrImageNotFound):
			status = http.StatusNotFound
		case errors.Is(err, repositories.ErrImageWrongProduct):
			status = http.StatusBadRequest
		case strings.Contains(err.Error(), "cloudinary"):
			status = http.StatusBadGateway
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Slika obrisana"})
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	value, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neispravan parametar " + name})
		return 0, err
	}
	return uint(value), nil
}
