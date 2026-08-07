package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cloudName, apiKey, apiSecret string) (*CloudinaryStorage, error) {
	client, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("neuspela inicijalizacija Cloudinary klijenta: %w", err)
	}
	return &CloudinaryStorage{client: client}, nil
}

func (s *CloudinaryStorage) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	result, err := s.client.Upload.Upload(ctx, input.Reader, uploader.UploadParams{
		Folder:   input.Folder,
		PublicID: input.PublicID,
	})
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
		Width:    result.Width,
		Height:   result.Height,
		Format:   result.Format,
		Bytes:    int64(result.Bytes),
	}, nil
}

func (s *CloudinaryStorage) Delete(ctx context.Context, publicID string) error {
	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}

// FakeStorage koristi se u testovima; ne poziva pravi Cloudinary.
type FakeStorage struct {
	mu sync.Mutex

	UploadFn func(ctx context.Context, input UploadInput) (*UploadResult, error)
	DeleteFn func(ctx context.Context, publicID string) error

	Uploaded []UploadResult
	Deleted  []string
}

func NewFakeStorage() *FakeStorage {
	return &FakeStorage{
		UploadFn: func(ctx context.Context, input UploadInput) (*UploadResult, error) {
			return &UploadResult{
				URL:      "https://fake.cloudinary.test/" + input.PublicID + ".jpg",
				PublicID: input.Folder + "/" + input.PublicID,
				Width:    800,
				Height:   600,
				Format:   "jpg",
				Bytes:    1024,
			}, nil
		},
		DeleteFn: func(ctx context.Context, publicID string) error {
			return nil
		},
	}
}

func (f *FakeStorage) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.UploadFn == nil {
		return nil, fmt.Errorf("upload nije konfigurisan")
	}
	result, err := f.UploadFn(ctx, input)
	if err != nil {
		return nil, err
	}
	f.Uploaded = append(f.Uploaded, *result)
	return result, nil
}

func (f *FakeStorage) Delete(ctx context.Context, publicID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.DeleteFn == nil {
		return fmt.Errorf("delete nije konfigurisan")
	}
	err := f.DeleteFn(ctx, publicID)
	if err != nil {
		return err
	}
	f.Deleted = append(f.Deleted, publicID)
	return nil
}
