package config

import (
	"os"
	"reflect"
	"testing"
)

func TestCORSAllowedOriginsMultiple(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://app.example.com ")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins, err := CORSAllowedOrigins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"http://localhost:3000", "https://app.example.com"}
	if !reflect.DeepEqual(origins, want) {
		t.Fatalf("got %v, want %v", origins, want)
	}
}

func TestCORSAllowedOriginsRejectsWildcard(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "*")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	if _, err := CORSAllowedOrigins(); err == nil {
		t.Fatal("expected error for wildcard origin")
	}
}
