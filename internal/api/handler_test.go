package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/service"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) *Handler {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://root:secret@localhost:5432/simple_ledger?sslmode=disable"
	}
	sqlDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	store := db.NewStore(sqlDB)
	ledger := service.NewLedgerService(store)
	return NewHandler(ledger, store)
}

func TestRegisterHandler_BadRequest(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	rw := httptest.NewRecorder()
	h.Register(rw, req)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestRegisterHandler_Success(t *testing.T) {
	h := setupTestHandler(t)
	_ = InitTokenAuth("fV7sliKV3qn657I60wEFtw/Auk/0bNU9zdp30wFzfDg=")

	// Use a unique email for each test run
	email := "testuser_" + uuid.New().String() + "@example.com"
	body := map[string]string{"email": email, "password": "testpassword123"}
	b, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	rw := httptest.NewRecorder()

	h.Register(rw, req)
	assert.Equal(t, http.StatusCreated, rw.Code)
}

// Add more handler tests as needed (mock dependencies for full coverage)
