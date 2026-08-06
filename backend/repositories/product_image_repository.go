package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"am-keramika-backend/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound       = errors.New("proizvod nije pronađen")
	ErrImageNotFound         = errors.New("slika nije pronađena")
	ErrImageWrongProduct     = errors.New("slika ne pripada proizvodu")
	ErrMaxImagesReached      = errors.New("proizvod već ima maksimalan broj slika")
	ErrProductHasImages      = errors.New("proizvod ima slike; uklonite slike prije deaktivacije")
	ErrInvalidReorderRequest = errors.New("neispravan reorder zahtjev")
)

func ProductExists(productID uint) error {
	var count int64
	if err := database.DB.Model(&models.Product{}).Where("id = ?", productID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrProductNotFound
	}
	return nil
}

func CountProductImages(productID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.ProductImage{}).Where("product_id = ?", productID).Count(&count).Error
	return count, err
}

func ProductHasImages(productID uint) (bool, error) {
	count, err := CountProductImages(productID)
	return count > 0, err
}

func GetProductImages(productID uint) ([]models.ProductImage, error) {
	var images []models.ProductImage
	err := database.DB.Where("product_id = ?", productID).
		Order("is_primary DESC, sort_order ASC, id ASC").
		Find(&images).Error
	return images, err
}

func GetPrimaryImagesForProducts(productIDs []uint) (map[uint]models.ProductImage, error) {
	result := make(map[uint]models.ProductImage)
	if len(productIDs) == 0 {
		return result, nil
	}

	var images []models.ProductImage
	err := database.DB.Where("product_id IN ? AND is_primary = ?", productIDs, true).Find(&images).Error
	if err != nil {
		return nil, err
	}
	for _, image := range images {
		result[image.ProductID] = image
	}
	return result, nil
}

func GetProductImageByID(imageID uint) (*models.ProductImage, error) {
	var image models.ProductImage
	err := database.DB.First(&image, imageID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return &image, nil
}

type UploadedImageInput struct {
	URL      string
	PublicID string
	Width    int
	Height   int
	Format   string
	Bytes    int64
	IsPrimary bool
	SortOrder int
}

func CreateProductImages(productID uint, inputs []UploadedImageInput) ([]models.ProductImage, error) {
	images := make([]models.ProductImage, 0, len(inputs))
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, input := range inputs {
			image := models.ProductImage{
				ProductID: productID,
				URL:       input.URL,
				PublicID:  input.PublicID,
				IsPrimary: input.IsPrimary,
				SortOrder: input.SortOrder,
				Format:    input.Format,
			}
			if input.Width > 0 {
				w := input.Width
				image.Width = &w
			}
			if input.Height > 0 {
				h := input.Height
				image.Height = &h
			}
			if input.Bytes > 0 {
				b := input.Bytes
				image.Bytes = &b
			}
			if err := tx.Create(&image).Error; err != nil {
				return err
			}
			images = append(images, image)
		}
		return nil
	})
	return images, err
}

func UploadProductImages(ctx context.Context, store storage.ImageStorage, productID uint, validated []storage.ValidatedImage) ([]models.ProductImage, error) {
	if err := ProductExists(productID); err != nil {
		return nil, err
	}

	existingCount, err := CountProductImages(productID)
	if err != nil {
		return nil, err
	}
	if existingCount+int64(len(validated)) > storage.MaxImagesPerProduct {
		return nil, ErrMaxImagesReached
	}

	var hasPrimary int64
	if err := database.DB.Model(&models.ProductImage{}).
		Where("product_id = ? AND is_primary = ?", productID, true).
		Count(&hasPrimary).Error; err != nil {
		return nil, err
	}

	var maxSort int
	database.DB.Model(&models.ProductImage{}).
		Where("product_id = ?", productID).
		Select("COALESCE(MAX(sort_order), -1)").
		Scan(&maxSort)

	folder := storage.ProductImageFolder(productID)
	uploaded := make([]storage.UploadResult, 0, len(validated))
	uploadedPublicIDs := make([]string, 0, len(validated))

	cleanup := func() {
		for _, publicID := range uploadedPublicIDs {
			_ = store.Delete(ctx, publicID)
		}
	}

	dbInputs := make([]UploadedImageInput, 0, len(validated))
	for i, file := range validated {
		publicID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), uuid.NewString())
		result, err := store.Upload(ctx, storage.UploadInput{
			Reader:   file.Reader,
			Folder:   folder,
			PublicID: publicID,
		})
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("cloudinary upload neuspješan: %w", err)
		}
		uploaded = append(uploaded, *result)
		uploadedPublicIDs = append(uploadedPublicIDs, result.PublicID)

		isPrimary := existingCount == 0 && hasPrimary == 0 && i == 0
		dbInputs = append(dbInputs, UploadedImageInput{
			URL:       result.URL,
			PublicID:  result.PublicID,
			Width:     result.Width,
			Height:    result.Height,
			Format:    result.Format,
			Bytes:     result.Bytes,
			IsPrimary: isPrimary,
			SortOrder: maxSort + i + 1,
		})
	}

	images, err := CreateProductImages(productID, dbInputs)
	if err != nil {
		cleanup()
		return nil, err
	}
	return images, nil
}

func SetPrimaryProductImage(productID, imageID uint) (*models.ProductImage, error) {
	image, err := GetProductImageByID(imageID)
	if err != nil {
		return nil, err
	}
	if image.ProductID != productID {
		return nil, ErrImageWrongProduct
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ProductImage{}).
			Where("product_id = ?", productID).
			Update("is_primary", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.ProductImage{}).
			Where("id = ? AND product_id = ?", imageID, productID).
			Update("is_primary", true).Error
	})
	if err != nil {
		return nil, err
	}

	return GetProductImageByID(imageID)
}

func ReorderProductImages(productID uint, imageIDs []uint) ([]models.ProductImage, error) {
	images, err := GetProductImages(productID)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return images, nil
	}
	if len(imageIDs) != len(images) {
		return nil, ErrInvalidReorderRequest
	}

	existing := make(map[uint]models.ProductImage, len(images))
	for _, image := range images {
		existing[image.ID] = image
	}

	seen := make(map[uint]struct{}, len(imageIDs))
	for _, id := range imageIDs {
		if _, ok := existing[id]; !ok {
			return nil, ErrInvalidReorderRequest
		}
		if _, dup := seen[id]; dup {
			return nil, ErrInvalidReorderRequest
		}
		seen[id] = struct{}{}
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for order, id := range imageIDs {
			if err := tx.Model(&models.ProductImage{}).
				Where("id = ? AND product_id = ?", id, productID).
				Update("sort_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return GetProductImages(productID)
}

func DeleteProductImage(ctx context.Context, store storage.ImageStorage, productID, imageID uint) error {
	image, err := GetProductImageByID(imageID)
	if err != nil {
		return err
	}
	if image.ProductID != productID {
		return ErrImageWrongProduct
	}

	if err := store.Delete(ctx, image.PublicID); err != nil {
		return fmt.Errorf("cloudinary brisanje neuspješno: %w", err)
	}

	wasPrimary := image.IsPrimary
	if err := database.DB.Delete(&models.ProductImage{}, imageID).Error; err != nil {
		return err
	}

	if !wasPrimary {
		return nil
	}

	var next models.ProductImage
	err = database.DB.Where("product_id = ?", productID).
		Order("sort_order ASC, id ASC").
		First(&next).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ProductImage{}).
			Where("product_id = ?", productID).
			Update("is_primary", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.ProductImage{}).
			Where("id = ?", next.ID).
			Update("is_primary", true).Error
	})
}
