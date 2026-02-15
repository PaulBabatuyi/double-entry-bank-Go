package api

import (
	"log"
	"os"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var (
	TokenAuth *jwtauth.JWTAuth
	jwtSecret = []byte("JWT_SECRET")
)

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "jwttoken"
		// using hardcoded JWT_SECRET – DO NOT do this in production!
	}
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	TokenAuth = jwtauth.New("HS256", []byte(secret), nil)
}

// GenerateToken for login
func GenerateToken(userID uuid.UUID) (string, error) {
	claims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	_, tokenString, err := TokenAuth.Encode(claims)
	return tokenString, err
}
