package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var errEmptySecret = errors.New("SECRET_KEY is not configured")

// GenerateToken creates a signed JWT for the given login.
func GenerateToken(login string) (string, error) {
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		return "", errEmptySecret
	}

	claims := jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken verifies a JWT and returns nil if valid.
func ValidateToken(tokenString string) error {
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		return errEmptySecret
	}

	tok, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Only HS384 is allowed — not HS256 or HS512.
		if token.Method != jwt.SigningMethodHS384 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}

	if !tok.Valid {
		return errors.New("invalid token")
	}

	// Require exp claim.
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}
	if _, hasExp := claims["exp"]; !hasExp {
		return errors.New("token missing exp claim")
	}

	return nil
}
