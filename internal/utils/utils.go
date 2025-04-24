package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/config"
	"github.com/dgrijalva/jwt-go"
)

// GenerateJWT generates a JWT token
func GenerateJWT(userID string, role string, cfg *config.Config) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Second * time.Duration(cfg.JWT.ExpiresIn)).Unix()

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token
func ValidateJWT(tokenString string, cfg *config.Config) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// CalculatePoints calculates points based on topup amount
func CalculatePoints(amount float64) int {
	switch {
	case amount >= 100 && amount < 200:
		return 1
	case amount >= 200 && amount < 300:
		return 2
	case amount >= 300 && amount < 400:
		return 3
	case amount >= 400 && amount < 500:
		return 4
	case amount >= 500 && amount < 1000:
		return 5
	case amount >= 1000:
		return 10
	default:
		return 0
	}
}

// GetDefaultEligibleDigits returns the default eligible digits for a given day of the week
func GetDefaultEligibleDigits(dayOfWeek time.Weekday) []int {
	switch dayOfWeek {
	case time.Monday:
		return []int{0, 1}
	case time.Tuesday:
		return []int{2, 3}
	case time.Wednesday:
		return []int{4, 5}
	case time.Thursday:
		return []int{6, 7}
	case time.Friday:
		return []int{8, 9}
	case time.Saturday:
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	default:
		return []int{}
	}
}
