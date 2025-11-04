package crypto

import (
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

func Hash(source string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(source), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func GenJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"role": role,
		"user_id": userID,
		"exp": time.Now().Add(24*time.Hour).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
}
