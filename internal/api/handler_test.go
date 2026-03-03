package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/service"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	testDB      *sql.DB
	testStore   *db.Store
	testHandler *Handler
	testLedger  *service.LedgerService
)

func TestMain(m *testing.M) {
	// Initialize JWT auth for testing
	err := InitTokenAuth("test-secret-key-with-at-least-32-characters-for-security")
	if err != nil {
		fmt.Printf("Failed to initialize token auth: %v\n", err)
		os.Exit(1)
	}

	// Setup test database connection
	connStr := os.Getenv("TEST_DB_URL")
	if connStr == "" {
		connStr = "postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable"
	}

	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	if err := testDB.Ping(); err != nil {
		fmt.Printf("Failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	testStore = db.NewStore(testDB)
	testLedger = service.NewLedgerService(testStore)
	testHandler = NewHandler(testLedger, testStore)

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

func createTestUser(t *testing.T, email, password string) (uuid.UUID, string) {
	t.Helper()
	ctx := context.Background()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := testStore.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          email,
		HashedPassword: string(hashed),
	})
	require.NoError(t, err)

	token, err := GenerateToken(user.ID)
	require.NoError(t, err)

	return user.ID, token
}

func cleanupTestUser(t *testing.T, userID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// Delete user's accounts and entries
	_, err := testDB.ExecContext(ctx, `
		DELETE FROM entries WHERE account_id IN (SELECT id FROM accounts WHERE owner_id = $1)
	`, userID)
	if err != nil {
		t.Logf("Warning: failed to delete entries: %v", err)
	}

	_, err = testDB.ExecContext(ctx, "DELETE FROM accounts WHERE owner_id = $1", userID)
	if err != nil {
		t.Logf("Warning: failed to delete accounts: %v", err)
	}

	_, err = testDB.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		t.Logf("Warning: failed to delete user: %v", err)
	}
}

func createTestAccount(t *testing.T, userID uuid.UUID, name string, initialBalance string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	acc, err := testStore.Queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		OwnerID:  uuid.NullUUID{UUID: userID, Valid: true},
		Name:     name,
		Currency: "NGN",
		IsSystem: false,
	})
	require.NoError(t, err)

	if initialBalance != "0" {
		bal, err := decimal.NewFromString(initialBalance)
		require.NoError(t, err)

		if bal.GreaterThan(decimal.Zero) {
			_, err = testStore.Queries.CreateEntry(ctx, sqlc.CreateEntryParams{
				AccountID:     acc.ID,
				Debit:         decimal.Zero.StringFixed(4),
				Credit:        bal.StringFixed(4),
				TransactionID: uuid.New(),
				OperationType: "deposit",
				Description:   sql.NullString{String: "Initial balance", Valid: true},
			})
			require.NoError(t, err)

			err = testStore.Queries.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
				Balance: bal.StringFixed(4),
				ID:      acc.ID,
			})
			require.NoError(t, err)
		}
	}

	return acc.ID
}

func TestHandler_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		email := fmt.Sprintf("test.%s@example.com", uuid.New().String()[:8])
		reqBody := map[string]string{
			"email":    email,
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp RegisterResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.UserID)
		assert.Equal(t, email, resp.Email)
		assert.NotEmpty(t, resp.Token)

		// Cleanup
		userID, _ := uuid.Parse(resp.UserID)
		cleanupTestUser(t, userID)
	})

	t.Run("missing email", func(t *testing.T) {
		reqBody := map[string]string{
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing password", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "test@example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("duplicate email", func(t *testing.T) {
		email := fmt.Sprintf("duplicate.%s@example.com", uuid.New().String()[:8])
		userID, _ := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		reqBody := map[string]string{
			"email":    email,
			"password": "password456",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestHandler_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		email := fmt.Sprintf("login.%s@example.com", uuid.New().String()[:8])
		password := "password123"
		userID, _ := createTestUser(t, email, password)
		defer cleanupTestUser(t, userID)

		reqBody := map[string]string{
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp TokenResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
	})

	t.Run("invalid email", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		email := fmt.Sprintf("wrongpass.%s@example.com", uuid.New().String()[:8])
		userID, _ := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		reqBody := map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		testHandler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandler_CreateAccount(t *testing.T) {
	t.Run("successful account creation", func(t *testing.T) {
		email := fmt.Sprintf("createacc.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		reqBody := map[string]string{
			"name": "Savings Account",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)

		// Add JWT to context
		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.CreateAccount(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp AccountResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "Savings Account", resp.Name)
		assert.Equal(t, "0.0000", resp.Balance)
	})

	t.Run("missing name", func(t *testing.T) {
		email := fmt.Sprintf("noname.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		reqBody := map[string]string{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.CreateAccount(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_Deposit(t *testing.T) {
	t.Run("successful deposit", func(t *testing.T) {
		email := fmt.Sprintf("deposit.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		accountID := createTestAccount(t, userID, "Test Account", "0")

		reqBody := map[string]string{
			"amount": "1000.0000",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/accounts/%s/deposit", accountID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.Deposit(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		// Create two users
		email1 := fmt.Sprintf("user1.%s@example.com", uuid.New().String()[:8])
		userID1, _ := createTestUser(t, email1, "password123")
		defer cleanupTestUser(t, userID1)

		email2 := fmt.Sprintf("user2.%s@example.com", uuid.New().String()[:8])
		userID2, token2 := createTestUser(t, email2, "password123")
		defer cleanupTestUser(t, userID2)

		// User1 creates account
		accountID := createTestAccount(t, userID1, "User1 Account", "0")

		// User2 tries to deposit to User1's account
		reqBody := map[string]string{
			"amount": "100.0000",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/accounts/%s/deposit", accountID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token2)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token2)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.Deposit(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestHandler_Withdraw(t *testing.T) {
	t.Run("successful withdrawal", func(t *testing.T) {
		email := fmt.Sprintf("withdraw.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		accountID := createTestAccount(t, userID, "Test Account", "1000.0000")

		reqBody := map[string]string{
			"amount": "300.0000",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/accounts/%s/withdraw", accountID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.Withdraw(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		email := fmt.Sprintf("insufficient.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		accountID := createTestAccount(t, userID, "Test Account", "100.0000")

		reqBody := map[string]string{
			"amount": "200.0000",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/accounts/%s/withdraw", accountID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.Withdraw(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_Transfer(t *testing.T) {
	t.Run("successful transfer", func(t *testing.T) {
		email := fmt.Sprintf("transfer.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		fromID := createTestAccount(t, userID, "From Account", "1000.0000")
		toID := createTestAccount(t, userID, "To Account", "500.0000")

		reqBody := map[string]interface{}{
			"from_id": fromID.String(),
			"to_id":   toID.String(),
			"amount":  "300.0000",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.Transfer(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_GetAccount(t *testing.T) {
	t.Run("successful get account", func(t *testing.T) {
		email := fmt.Sprintf("getacc.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		accountID := createTestAccount(t, userID, "My Account", "1500.0000")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/accounts/%s", accountID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.GetAccount(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AccountResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, accountID.String(), resp.ID)
		assert.Equal(t, "My Account", resp.Name)
		assert.Equal(t, "1500.0000", resp.Balance)
	})
}

func TestHandler_ListAccounts(t *testing.T) {
	t.Run("list user accounts", func(t *testing.T) {
		email := fmt.Sprintf("listacc.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		createTestAccount(t, userID, "Account 1", "100.0000")
		createTestAccount(t, userID, "Account 2", "200.0000")

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.ListAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []AccountResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(resp), 2)
	})
}

func TestHandler_ReconcileAccount(t *testing.T) {
	t.Run("reconcile account", func(t *testing.T) {
		email := fmt.Sprintf("reconcile.%s@example.com", uuid.New().String()[:8])
		userID, token := createTestUser(t, email, "password123")
		defer cleanupTestUser(t, userID)

		accountID := createTestAccount(t, userID, "Reconcile Account", "0")

		// Perform deposit
		err := testLedger.Deposit(context.Background(), accountID, "500.0000")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/accounts/%s/reconcile", accountID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := req.Context()
		jwtToken, _ := TokenAuth.Decode(token)
		ctx = jwtauth.NewContext(ctx, jwtToken, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", accountID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		testHandler.ReconcileAccount(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp ReconcileResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Matched)
	})
}
