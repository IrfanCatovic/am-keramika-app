package storage

import (
	"context"
	"io"
)

type UploadInput struct {
	Reader   io.Reader
	Folder   string
	PublicID string
}

type UploadResult struct {
	URL      string
	PublicID string
	Width    int
	Height   int
	Format   string
	Bytes    int64
}

type ImageStorage interface {
	Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
	Delete(ctx context.Context, publicID string) error
}
