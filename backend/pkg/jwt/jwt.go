package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GenerateAccessToken(userID pgtype.UUID, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
		"type":    "access",
	})

	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return s, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func ParseAccessToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("Invalid token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return "", errors.New("Not an access token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("Invalid user id in token")
	}

	return userID, nil
}
