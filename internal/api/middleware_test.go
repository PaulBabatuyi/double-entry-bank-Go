package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitTokenAuth(t *testing.T) {
	t.Run("successful initialization with valid secret", func(t *testing.T) {
		err := InitTokenAuth("this-is-a-very-long-secret-key-for-testing-jwt-tokens")
		require.NoError(t, err)
		assert.NotNil(t, TokenAuth)
	})

	t.Run("error with empty secret", func(t *testing.T) {
		err := InitTokenAuth("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("error with short secret", func(t *testing.T) {
		err := InitTokenAuth("short")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 32 characters")
	})
}

func TestInitTokenAuthFromEnv(t *testing.T) {
	t.Run("error when JWT_SECRET not set", func(t *testing.T) {
		// This test assumes JWT_SECRET is not set in environment
		// In actual test environment, this might need adjustment
		t.Setenv("JWT_SECRET", "")
		err := InitTokenAuthFromEnv()
		assert.Error(t, err)
	})
}

func TestGenerateToken(t *testing.T) {
	// Initialize token auth first
	err := InitTokenAuth("test-secret-key-with-at-least-32-characters-for-testing")
	require.NoError(t, err)

	t.Run("generates valid token", func(t *testing.T) {
		userID := uuid.New()
		token, err := GenerateToken(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be decoded
		jwtToken, err := TokenAuth.Decode(token)
		require.NoError(t, err)
		assert.NotNil(t, jwtToken)

		// Verify claims via jwx v3 Get(key, *dst) API
		var userIDVal string
		claimErr := jwtToken.Get("user_id", &userIDVal)
		require.NoError(t, claimErr)
		assert.Equal(t, userID.String(), userIDVal)

		expTime, hasExp := jwtToken.Expiration()
		assert.True(t, hasExp)
		assert.False(t, expTime.IsZero())
	})

	t.Run("error when token auth not initialized", func(t *testing.T) {
		// Save current TokenAuth
		savedAuth := TokenAuth
		TokenAuth = nil

		userID := uuid.New()
		_, err := GenerateToken(userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")

		// Restore TokenAuth
		TokenAuth = savedAuth
	})

	t.Run("generates different tokens for different users", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()

		token1, err := GenerateToken(userID1)
		require.NoError(t, err)

		token2, err := GenerateToken(userID2)
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})
}

func TestRespondJSON(t *testing.T) {
	// This is tested through handler tests
	// Just ensuring the function signature is correct
	t.Run("function exists and can be called", func(t *testing.T) {
		// This test verifies compilation more than runtime behavior
		// Actual behavior is tested in handler tests
		assert.NotNil(t, respondJSON)
		assert.NotNil(t, respondError)
	})
}

func TestToAccountResponse(t *testing.T) {
	// Initialize for testing
	err := InitTokenAuth("test-secret-key-with-at-least-32-characters-for-testing")
	require.NoError(t, err)

	t.Run("mapper functions are available", func(t *testing.T) {
		// Verify mapper functions compile and are accessible (behaviour tested via handler tests)
		assert.NotNil(t, toAccountResponse)
		assert.NotNil(t, toEntryResponse)
	})
}
