package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(login string) (string, error) {
	claims := jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)

	secret := []byte(os.Getenv("SECRET_KEY"))

	return token.SignedString(secret)
}

func ValidateToken(tokenString string) error {
	secret := []byte(os.Getenv("SECRET_KEY"))

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		println(err.Error())
	}

	return err
}
