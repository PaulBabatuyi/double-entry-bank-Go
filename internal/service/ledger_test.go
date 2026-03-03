package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB    *sql.DB
	testStore *db.Store
)

func TestMain(m *testing.M) {
	// Setup test database connection
	connStr := os.Getenv("TEST_DB_URL")
	if connStr == "" {
		connStr = "postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable"
	}

	var err error
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

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

func setupTestAccount(t *testing.T, name string, initialBalance string) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	acc, err := testStore.Queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		OwnerID:  uuid.NullUUID{Valid: false},
		Name:     name,
		Currency: "NGN",
		IsSystem: false,
	})
	require.NoError(t, err)

	if initialBalance != "0" {
		// Add initial balance via entry
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

func cleanupTestAccount(t *testing.T, accountID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// Delete entries first (foreign key constraint)
	_, err := testDB.ExecContext(ctx, "DELETE FROM entries WHERE account_id = $1", accountID)
	if err != nil {
		t.Logf("Warning: failed to delete entries: %v", err)
	}

	// Delete account
	_, err = testDB.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1 AND is_system = false", accountID)
	if err != nil {
		t.Logf("Warning: failed to delete account: %v", err)
	}
}

func getAccountBalance(t *testing.T, accountID uuid.UUID) string {
	t.Helper()
	ctx := context.Background()

	acc, err := testStore.Queries.GetAccount(ctx, accountID)
	require.NoError(t, err)

	return acc.Balance
}

func TestLedgerService_Deposit(t *testing.T) {
	t.Run("successful deposit", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Deposit Account", "0")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Deposit(ctx, accountID, "1000.0000")
		require.NoError(t, err)

		balance := getAccountBalance(t, accountID)
		assert.Equal(t, "1000.0000", balance)
	})

	t.Run("multiple deposits accumulate", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Multi Deposit", "0")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Deposit(ctx, accountID, "500.0000")
		require.NoError(t, err)

		err = ledger.Deposit(ctx, accountID, "300.5000")
		require.NoError(t, err)

		balance := getAccountBalance(t, accountID)
		assert.Equal(t, "800.5000", balance)
	})

	t.Run("invalid amount - negative", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Invalid Deposit", "0")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Deposit(ctx, accountID, "-100.0000")
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("invalid amount - zero", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Zero Deposit", "0")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Deposit(ctx, accountID, "0")
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("invalid amount - non-numeric", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Bad Deposit", "0")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Deposit(ctx, accountID, "abc")
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("non-existent account", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		randomID := uuid.New()
		err := ledger.Deposit(ctx, randomID, "100.0000")
		assert.Error(t, err)
	})
}

func TestLedgerService_Withdraw(t *testing.T) {
	t.Run("successful withdrawal", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Withdraw Account", "1000.0000")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Withdraw(ctx, accountID, "300.0000")
		require.NoError(t, err)

		balance := getAccountBalance(t, accountID)
		assert.Equal(t, "700.0000", balance)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Insufficient Funds", "100.0000")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Withdraw(ctx, accountID, "200.0000")
		assert.ErrorIs(t, err, ErrInsufficientFunds)

		// Balance should remain unchanged
		balance := getAccountBalance(t, accountID)
		assert.Equal(t, "100.0000", balance)
	})

	t.Run("exact balance withdrawal", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Exact Withdrawal", "500.0000")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Withdraw(ctx, accountID, "500.0000")
		require.NoError(t, err)

		balance := getAccountBalance(t, accountID)
		assert.Equal(t, "0.0000", balance)
	})

	t.Run("invalid amount", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Invalid Withdrawal", "100.0000")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Withdraw(ctx, accountID, "-50.0000")
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}

func TestLedgerService_Transfer(t *testing.T) {
	t.Run("successful transfer", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		fromID := setupTestAccount(t, "Test From Account", "1000.0000")
		defer cleanupTestAccount(t, fromID)

		toID := setupTestAccount(t, "Test To Account", "500.0000")
		defer cleanupTestAccount(t, toID)

		err := ledger.Transfer(ctx, fromID, toID, "300.0000")
		require.NoError(t, err)

		fromBalance := getAccountBalance(t, fromID)
		toBalance := getAccountBalance(t, toID)

		assert.Equal(t, "700.0000", fromBalance)
		assert.Equal(t, "800.0000", toBalance)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		fromID := setupTestAccount(t, "Test Transfer From", "100.0000")
		defer cleanupTestAccount(t, fromID)

		toID := setupTestAccount(t, "Test Transfer To", "0")
		defer cleanupTestAccount(t, toID)

		err := ledger.Transfer(ctx, fromID, toID, "200.0000")
		assert.ErrorIs(t, err, ErrInsufficientFunds)

		// Balances should remain unchanged
		fromBalance := getAccountBalance(t, fromID)
		toBalance := getAccountBalance(t, toID)

		assert.Equal(t, "100.0000", fromBalance)
		assert.Equal(t, "0.0000", toBalance)
	})

	t.Run("same account transfer", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Same Account", "1000.0000")
		defer cleanupTestAccount(t, accountID)

		err := ledger.Transfer(ctx, accountID, accountID, "100.0000")
		assert.ErrorIs(t, err, ErrSameAccountTransfer)
	})

	t.Run("invalid amount", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		fromID := setupTestAccount(t, "Test Invalid Transfer From", "1000.0000")
		defer cleanupTestAccount(t, fromID)

		toID := setupTestAccount(t, "Test Invalid Transfer To", "0")
		defer cleanupTestAccount(t, toID)

		err := ledger.Transfer(ctx, fromID, toID, "0")
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}

func TestLedgerService_ReconcileAccount(t *testing.T) {
	t.Run("balanced account reconciles", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Reconcile Account", "0")
		defer cleanupTestAccount(t, accountID)

		// Perform some operations
		err := ledger.Deposit(ctx, accountID, "1000.0000")
		require.NoError(t, err)

		err = ledger.Withdraw(ctx, accountID, "300.0000")
		require.NoError(t, err)

		// Reconcile
		matched, err := ledger.ReconcileAccount(ctx, accountID)
		require.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("fresh account reconciles", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Fresh Account", "0")
		defer cleanupTestAccount(t, accountID)

		matched, err := ledger.ReconcileAccount(ctx, accountID)
		require.NoError(t, err)
		assert.True(t, matched)
	})
}

func TestLedgerService_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent deposits should not cause race", func(t *testing.T) {
		ledger := NewLedgerService(testStore)
		ctx := context.Background()

		accountID := setupTestAccount(t, "Test Concurrent Account", "0")
		defer cleanupTestAccount(t, accountID)

		// Run deposits concurrently
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				err := ledger.Deposit(ctx, accountID, "100.0000")
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all deposits
		for i := 0; i < 5; i++ {
			<-done
		}

		// Balance should be 500
		balance := getAccountBalance(t, accountID)
		assert.Equal(t, "500.0000", balance)

		// Reconcile should pass
		matched, err := ledger.ReconcileAccount(ctx, accountID)
		require.NoError(t, err)
		assert.True(t, matched)
	})
}

func TestValidatePositiveAmount(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		errType   error
		wantValue string
	}{
		{
			name:      "valid positive amount",
			input:     "100.5000",
			wantErr:   false,
			wantValue: "100.5000",
		},
		{
			name:    "zero amount",
			input:   "0",
			wantErr: true,
			errType: ErrInvalidAmount,
		},
		{
			name:    "negative amount",
			input:   "-50.0000",
			wantErr: true,
			errType: ErrInvalidAmount,
		},
		{
			name:    "invalid string",
			input:   "abc",
			wantErr: true,
			errType: ErrInvalidAmount,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errType: ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePositiveAmount(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, got.StringFixed(4))
			}
		})
	}
}
