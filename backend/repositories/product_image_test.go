package repositories

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"am-keramika-backend/storage"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var jpegBytes = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

func setupImageTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.ProductImage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedImageTestProduct(t *testing.T) models.Product {
	t.Helper()
	category := models.Category{Name: "Keramika", Slug: "keramika", IsActive: true}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("category: %v", err)
	}
	product := models.Product{
		Name: "Proizvod", Slug: "proizvod", CategoryID: category.ID,
		Unit: "kom", SalePrice: 10, IsActive: true,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}
	return product
}

func validatedJPEG() storage.ValidatedImage {
	return storage.ValidatedImage{
		Reader:   bytes.NewReader(jpegBytes),
		MIMEType: "image/jpeg",
		Size:     int64(len(jpegBytes)),
	}
}

func TestUploadProductImagesFirstBecomesPrimary(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()

	images, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if len(images) != 1 || !images[0].IsPrimary {
		t.Fatalf("expected primary first image, got %+v", images)
	}
}

func TestUploadProductImagesAdditionalDoesNotChangePrimary(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()

	first, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}

	secondBatch, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	if secondBatch[0].IsPrimary {
		t.Fatal("second image must not become primary")
	}

	reloaded, _ := GetProductImageByID(first[0].ID)
	if !reloaded.IsPrimary {
		t.Fatal("original primary must remain")
	}
}

func TestUploadProductImagesProductNotFound(t *testing.T) {
	setupImageTestDB(t)
	fake := storage.NewFakeStorage()
	_, err := UploadProductImages(context.Background(), fake, 999, []storage.ValidatedImage{validatedJPEG()})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestUploadProductImagesMaxEight(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()

	for i := 0; i < 8; i++ {
		if _, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()}); err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
	}
	_, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})
	if !errors.Is(err, ErrMaxImagesReached) {
		t.Fatalf("expected max images error, got %v", err)
	}
}

func TestSetPrimaryProductImage(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	images, _ := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG(), validatedJPEG()})

	updated, err := SetPrimaryProductImage(product.ID, images[1].ID)
	if err != nil {
		t.Fatalf("set primary: %v", err)
	}
	if !updated.IsPrimary {
		t.Fatal("expected selected image primary")
	}
	first, _ := GetProductImageByID(images[0].ID)
	if first.IsPrimary {
		t.Fatal("previous primary must be cleared")
	}
}

func TestSetPrimaryWrongProduct(t *testing.T) {
	setupImageTestDB(t)
	p1 := seedImageTestProduct(t)
	category := models.Category{Name: "Druga", Slug: "druga", IsActive: true}
	database.DB.Create(&category)
	p2 := models.Product{Name: "Drugi", Slug: "drugi", CategoryID: category.ID, Unit: "kom", SalePrice: 10, IsActive: true}
	database.DB.Create(&p2)

	fake := storage.NewFakeStorage()
	images, _ := UploadProductImages(context.Background(), fake, p1.ID, []storage.ValidatedImage{validatedJPEG()})

	_, err := SetPrimaryProductImage(p2.ID, images[0].ID)
	if !errors.Is(err, ErrImageWrongProduct) {
		t.Fatalf("expected wrong product, got %v", err)
	}
}

func TestReorderProductImages(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	images, _ := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG(), validatedJPEG(), validatedJPEG()})

	reordered, err := ReorderProductImages(product.ID, []uint{images[2].ID, images[0].ID, images[1].ID})
	if err != nil {
		t.Fatalf("reorder: %v", err)
	}

	sortOrders := map[uint]int{}
	for _, image := range reordered {
		sortOrders[image.ID] = image.SortOrder
	}
	if sortOrders[images[2].ID] != 0 || sortOrders[images[0].ID] != 1 || sortOrders[images[1].ID] != 2 {
		t.Fatalf("unexpected sort orders: %+v", sortOrders)
	}
}

func TestReorderDuplicateRejected(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	images, _ := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG(), validatedJPEG()})

	_, err := ReorderProductImages(product.ID, []uint{images[0].ID, images[0].ID})
	if !errors.Is(err, ErrInvalidReorderRequest) {
		t.Fatalf("expected invalid reorder, got %v", err)
	}
}

func TestDeletePrimaryPromotesNext(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	images, _ := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG(), validatedJPEG()})

	if err := DeleteProductImage(context.Background(), fake, product.ID, images[0].ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	remaining, _ := GetProductImages(product.ID)
	if len(remaining) != 1 || !remaining[0].IsPrimary {
		t.Fatalf("expected one remaining primary, got %+v", remaining)
	}
}

func TestUploadFailureDoesNotCreateDBRows(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	call := 0
	fake.UploadFn = func(ctx context.Context, input storage.UploadInput) (*storage.UploadResult, error) {
		call++
		if call == 2 {
			return nil, errors.New("upload failed")
		}
		return &storage.UploadResult{URL: "https://x/y", PublicID: "x/y", Format: "jpg", Bytes: 10}, nil
	}

	_, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG(), validatedJPEG()})
	if err == nil {
		t.Fatal("expected upload error")
	}
	count, _ := CountProductImages(product.ID)
	if count != 0 {
		t.Fatalf("expected 0 db rows after failed batch, got %d", count)
	}
	if len(fake.Deleted) == 0 {
		t.Fatal("expected cloudinary cleanup")
	}
}

func TestDBFailureTriggersCleanup(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()

	existing, err := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})
	if err != nil {
		t.Fatalf("seed image: %v", err)
	}

	fake.UploadFn = func(ctx context.Context, input storage.UploadInput) (*storage.UploadResult, error) {
		return &storage.UploadResult{
			URL:      "https://x/y",
			PublicID: existing[0].PublicID,
			Format:   "jpg",
			Bytes:    10,
		}, nil
	}

	_, err = UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})
	if err == nil {
		t.Fatal("expected db failure on duplicate public id")
	}
	if len(fake.Deleted) == 0 {
		t.Fatal("expected cleanup after db failure")
	}
}

func TestDeleteCloudinaryFailureKeepsDBRow(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	images, _ := UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})

	fake.DeleteFn = func(ctx context.Context, publicID string) error {
		return errors.New("cloudinary down")
	}
	err := DeleteProductImage(context.Background(), fake, product.ID, images[0].ID)
	if err == nil || !strings.Contains(err.Error(), "cloudinary") {
		t.Fatalf("expected cloudinary error, got %v", err)
	}
	count, _ := CountProductImages(product.ID)
	if count != 1 {
		t.Fatalf("db row must remain, count=%d", count)
	}
}

func TestDeactivateProductWithImagesConflict(t *testing.T) {
	setupImageTestDB(t)
	product := seedImageTestProduct(t)
	fake := storage.NewFakeStorage()
	_, _ = UploadProductImages(context.Background(), fake, product.ID, []storage.ValidatedImage{validatedJPEG()})

	err := DeactivateProduct(strconv.FormatUint(uint64(product.ID), 10))
	if !errors.Is(err, ErrProductHasImages) {
		t.Fatalf("expected conflict err, got %v", err)
	}
}
