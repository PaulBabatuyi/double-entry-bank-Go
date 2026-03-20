package api

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitTokenAuthFromEnv_MissingSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	err := InitTokenAuthFromEnv()
	assert.Error(t, err)
}

func TestInitTokenAuth_Success(t *testing.T) {
	// Use a valid 32+ char secret
	secret := "fV7sliKV3qn657I60wEFtw/Auk/0bNU9zdp30wFzfDg="
	err := InitTokenAuth(secret)
	assert.NoError(t, err)
}
