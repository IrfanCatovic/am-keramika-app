package auth

import (
	"os"
	"testing"
	"time"

	"am-keramika-backend/models"

	"github.com/golang-jwt/jwt/v5"
)

func TestNormalizeUsername(t *testing.T) {
	if got := NormalizeUsername("  SeF  "); got != "sef" {
		t.Fatalf("expected sef, got %q", got)
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !CheckPassword(hash, "password123") {
		t.Fatal("expected password match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-jwt")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateToken(1, "sef", models.RoleBoss, 3)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != 1 || claims.Username != "sef" || claims.Role != models.RoleBoss || claims.TokenVersion != 3 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseInvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-jwt")
	defer os.Unsetenv("JWT_SECRET")

	if _, err := ParseToken("not.a.token"); err == nil {
		t.Fatal("expected invalid token error")
	}
}

func TestParseExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-jwt")
	defer os.Unsetenv("JWT_SECRET")

	now := time.Now()
	claims := Claims{
		UserID:   1,
		Username: "sef",
		Role:     models.RoleBoss,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret-for-jwt"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := ParseToken(signed); err == nil {
		t.Fatal("expected expired token error")
	}
}
