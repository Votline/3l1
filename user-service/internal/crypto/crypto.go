package crypto

import (
	"os"
	"time"
	"errors"

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
func CheckPswd(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type UserInfo struct {
	Role string
	UserID string
}
func GenJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"role": role,
		"user_id": userID,
		"exp": time.Now().Add(24*time.Hour).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
}
func ExtJWT(tokenString string) (UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error){
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid JWT token")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return UserInfo{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return UserInfo{}, errors.New("Failed to extract data from JWT token")
	}

	info := UserInfo{}
	if userID, ok := claims["user_id"].(string); !ok {
		return info, errors.New("Failed to extract user_id from JWT token")
	} else { info.UserID = userID }

	if role, ok := claims["role"].(string); !ok {
		return info, errors.New("Failed to extract role from JWT token")
	} else { info.Role = role }

	return info, nil
}
