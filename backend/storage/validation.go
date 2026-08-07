package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	MaxImagesPerProduct = 8
	MaxImageSizeBytes   = 10 * 1024 * 1024
)

var allowedMIMETypes = map[string]string{
	"image/jpeg": "jpeg",
	"image/png":  "png",
	"image/webp": "webp",
}

type ValidatedImage struct {
	Reader   io.Reader
	MIMEType string
	Size     int64
}

func ValidateImageFile(headerSize int64, reader io.Reader) (ValidatedImage, error) {
	if headerSize <= 0 {
		return ValidatedImage{}, errors.New("prazan fajl")
	}
	if headerSize > MaxImageSizeBytes {
		return ValidatedImage{}, fmt.Errorf("fajl prelazi maksimalnu veličinu od %d MB", MaxImageSizeBytes/(1024*1024))
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(reader, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return ValidatedImage{}, errors.New("neuspelo čitanje fajla")
	}
	if n == 0 {
		return ValidatedImage{}, errors.New("prazan fajl")
	}

	mimeType := http.DetectContentType(buf[:n])
	if _, ok := allowedMIMETypes[mimeType]; !ok {
		return ValidatedImage{}, errors.New("nedozvoljen MIME tip; dozvoljeni su JPEG, PNG i WebP")
	}

	combined := io.MultiReader(bytes.NewReader(buf[:n]), reader)
	return ValidatedImage{
		Reader:   io.LimitReader(combined, MaxImageSizeBytes+1),
		MIMEType: mimeType,
		Size:     headerSize,
	}, nil
}

func ProductImageFolder(productID uint) string {
	return fmt.Sprintf("am-keramika/products/%d", productID)
}
