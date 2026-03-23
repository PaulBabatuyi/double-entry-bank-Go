package api

import (
	"errors"
	"os"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var (
	// TokenAuth holds the JWT authenticator used by the API package.
	TokenAuth *jwtauth.JWTAuth
)

// InitTokenAuthFromEnv initializes JWT auth using the JWT_SECRET environment variable.
func InitTokenAuthFromEnv() error {
	// Keep bootstrap simple: this function is called once from main().
	secret := os.Getenv("JWT_SECRET")
	return InitTokenAuth(secret)
}

// InitTokenAuth initializes JWT auth with the provided secret.
func InitTokenAuth(secret string) error {
	// Fail fast if JWT configuration is insecure or missing.
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}

	if len(secret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}

	TokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	return nil
}

// GenerateToken creates a signed JWT for the given user ID.
func GenerateToken(userID uuid.UUID) (string, error) {
	if TokenAuth == nil {
		return "", errors.New("token auth is not initialized")
	}

	// Include user identity and expiry in signed JWT claims.
	claims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	_, tokenString, err := TokenAuth.Encode(claims)
	return tokenString, err
}
