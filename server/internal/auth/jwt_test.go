package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func setSecret(t *testing.T, val string) {
	t.Helper()
	os.Setenv("SECRET_KEY", val)
	t.Cleanup(func() { os.Unsetenv("SECRET_KEY") })
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	setSecret(t, "")

	_, err := GenerateToken("test")
	if err == nil {
		t.Fatal("expected error for empty SECRET_KEY")
	}
}

func TestGenerateToken_Valid(t *testing.T) {
	setSecret(t, "test-secret-key-123")

	token, err := GenerateToken("admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateToken_EmptySecret(t *testing.T) {
	setSecret(t, "")

	// Generate a token with a different secret to have something to validate
	_, _ = GenerateToken("admin") // creates with empty secret — will fail

	err := ValidateToken("some.token.here")
	if err == nil {
		t.Fatal("expected error for empty SECRET_KEY")
	}
}

func TestValidateToken_WrongSignature(t *testing.T) {
	setSecret(t, "correct-secret")

	// Create a token signed with a wrong secret
	wrongClaims := jwt.MapClaims{
		"login": "admin",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	wrongToken := jwt.NewWithClaims(jwt.SigningMethodHS384, wrongClaims)
	wrongSigned, _ := wrongToken.SignedString([]byte("wrong-secret"))

	err := ValidateToken(wrongSigned)
	if err == nil {
		t.Error("expected error for wrong signature")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	setSecret(t, "test-secret-key-123")

	expiredClaims := jwt.MapClaims{
		"login": "admin",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS384, expiredClaims)
	expiredSigned, _ := expiredToken.SignedString([]byte("test-secret-key-123"))

	err := ValidateToken(expiredSigned)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateToken_MissingExp(t *testing.T) {
	setSecret(t, "test-secret-key-123")

	// Create a token without exp claim
	claims := jwt.MapClaims{
		"login": "admin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, _ := token.SignedString([]byte("test-secret-key-123"))

	err := ValidateToken(signed)
	if err == nil {
		t.Error("expected error for token missing exp claim")
	}
}

func TestValidateToken_WrongAlgorithm(t *testing.T) {
	setSecret(t, "test-secret-key-123")

	// HS256 is a different HMAC algorithm — should be rejected
	// since we require HS384 specifically
	claims := jwt.MapClaims{
		"login": "admin",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("test-secret-key-123"))

	err := ValidateToken(signed)
	if err == nil {
		t.Error("expected error for HS256 algorithm (only HS384 allowed)")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	setSecret(t, "test-secret-key-123")

	token, err := GenerateToken("admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	err = ValidateToken(token)
	if err != nil {
		t.Errorf("expected valid token, got error: %v", err)
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	setSecret(t, "test-secret-key-123")

	err := ValidateToken("not-a-valid-jwt")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}
