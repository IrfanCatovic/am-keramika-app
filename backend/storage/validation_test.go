package storage

import (
	"bytes"
	"strings"
	"testing"
)

var jpegHeader = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}

func TestValidateImageFileAcceptsJPEG(t *testing.T) {
	_, err := ValidateImageFile(int64(len(jpegHeader)+100), bytes.NewReader(append(jpegHeader, make([]byte, 100)...)))
	if err != nil {
		t.Fatalf("expected jpeg to pass, got %v", err)
	}
}

func TestValidateImageFileRejectsInvalidMIME(t *testing.T) {
	_, err := ValidateImageFile(12, bytes.NewReader([]byte("not-an-image")))
	if err == nil || !strings.Contains(err.Error(), "MIME") {
		t.Fatalf("expected MIME error, got %v", err)
	}
}

func TestValidateImageFileRejectsOversized(t *testing.T) {
	_, err := ValidateImageFile(MaxImageSizeBytes+1, bytes.NewReader(jpegHeader))
	if err == nil || !strings.Contains(err.Error(), "10 MB") {
		t.Fatalf("expected size error, got %v", err)
	}
}
