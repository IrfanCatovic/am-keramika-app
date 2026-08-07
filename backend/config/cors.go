package config

import (
	"errors"
	"os"
	"strings"
)

func CORSAllowedOrigins() ([]string, error) {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("CORS_ALLOWED_ORIGINS nije postavljen")
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		if origin == "*" {
			return nil, errors.New("CORS_ALLOWED_ORIGINS ne sme sadržavati *")
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return nil, errors.New("CORS_ALLOWED_ORIGINS nije validan")
	}

	return origins, nil
}
