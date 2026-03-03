package api

import (
	"errors"
	"os"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var (
	TokenAuth *jwtauth.JWTAuth
)

func InitTokenAuthFromEnv() error {
	secret := os.Getenv("JWT_SECRET")
	return InitTokenAuth(secret)
}

func InitTokenAuth(secret string) error {
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}

	if len(secret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}

	TokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	return nil
}

// GenerateToken for login
func GenerateToken(userID uuid.UUID) (string, error) {
	if TokenAuth == nil {
		return "", errors.New("token auth is not initialized")
	}

	claims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	_, tokenString, err := TokenAuth.Encode(claims)
	return tokenString, err
}
