package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB    *sql.DB
	testStore *Store
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

	testStore = NewStore(testDB)

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

func setupTestAccount(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	acc, err := testStore.Queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		OwnerID:  uuid.NullUUID{Valid: false},
		Name:     "Test TX Account",
		Currency: "NGN",
		IsSystem: false,
	})
	require.NoError(t, err)

	return acc.ID
}

func cleanupTestAccount(t *testing.T, accountID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	_, err := testDB.ExecContext(ctx, "DELETE FROM entries WHERE account_id = $1", accountID)
	if err != nil {
		t.Logf("Warning: failed to delete entries: %v", err)
	}

	_, err = testDB.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1 AND is_system = false", accountID)
	if err != nil {
		t.Logf("Warning: failed to delete account: %v", err)
	}
}

func TestStore_ExecTx_Success(t *testing.T) {
	t.Run("successful transaction commits", func(t *testing.T) {
		ctx := context.Background()
		accountID := setupTestAccount(t)
		defer cleanupTestAccount(t, accountID)

		err := testStore.ExecTx(ctx, func(q *sqlc.Queries) error {
			// Create entry
			_, err := q.CreateEntry(ctx, sqlc.CreateEntryParams{
				AccountID:     accountID,
				Debit:         decimal.Zero.StringFixed(4),
				Credit:        "100.0000",
				TransactionID: uuid.New(),
				OperationType: "deposit",
				Description:   sql.NullString{String: "Test deposit", Valid: true},
			})
			if err != nil {
				return err
			}

			// Update balance
			return q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
				Balance: "100.0000",
				ID:      accountID,
			})
		})

		require.NoError(t, err)

		// Verify changes persisted
		acc, err := testStore.Queries.GetAccount(ctx, accountID)
		require.NoError(t, err)
		assert.Equal(t, "100.0000", acc.Balance)
	})
}

func TestStore_ExecTx_Rollback(t *testing.T) {
	t.Run("failed transaction rolls back", func(t *testing.T) {
		ctx := context.Background()
		accountID := setupTestAccount(t)
		defer cleanupTestAccount(t, accountID)

		// Get initial balance
		initialAcc, err := testStore.Queries.GetAccount(ctx, accountID)
		require.NoError(t, err)

		err = testStore.ExecTx(ctx, func(q *sqlc.Queries) error {
			// Update balance
			err := q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
				Balance: "100.0000",
				ID:      accountID,
			})
			if err != nil {
				return err
			}

			// Intentionally cause error
			return fmt.Errorf("intentional error to trigger rollback")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "intentional error")

		// Verify balance unchanged
		acc, err := testStore.Queries.GetAccount(ctx, accountID)
		require.NoError(t, err)
		assert.Equal(t, initialAcc.Balance, acc.Balance)
	})
}

func TestStore_ExecTx_ConcurrentTransactions(t *testing.T) {
	t.Run("serializable isolation prevents race conditions", func(t *testing.T) {
		ctx := context.Background()
		accountID := setupTestAccount(t)
		defer cleanupTestAccount(t, accountID)

		// Set initial balance
		err := testStore.Queries.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: "1000.0000",
			ID:      accountID,
		})
		require.NoError(t, err)

		// Run concurrent transactions
		done := make(chan error, 2)

		// Transaction 1: Add 100
		go func() {
			err := testStore.ExecTx(ctx, func(q *sqlc.Queries) error {
				acc, err := q.GetAccountForUpdate(ctx, accountID)
				if err != nil {
					return err
				}

				currentBalance, _ := decimal.NewFromString(acc.Balance)
				newBalance := currentBalance.Add(decimal.NewFromInt(100))

				return q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
					Balance: newBalance.Sub(currentBalance).StringFixed(4),
					ID:      accountID,
				})
			})
			done <- err
		}()

		// Transaction 2: Add 200
		go func() {
			err := testStore.ExecTx(ctx, func(q *sqlc.Queries) error {
				acc, err := q.GetAccountForUpdate(ctx, accountID)
				if err != nil {
					return err
				}

				currentBalance, _ := decimal.NewFromString(acc.Balance)
				newBalance := currentBalance.Add(decimal.NewFromInt(200))

				return q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
					Balance: newBalance.Sub(currentBalance).StringFixed(4),
					ID:      accountID,
				})
			})
			done <- err
		}()

		// Wait for both transactions
		err1 := <-done
		err2 := <-done

		// At least one should succeed (the other might get serialization error)
		// But the final balance should be correct
		if err1 != nil && err2 != nil {
			t.Fatalf("Both transactions failed: %v, %v", err1, err2)
		}

		// Verify final balance
		acc, err := testStore.Queries.GetAccount(ctx, accountID)
		require.NoError(t, err)

		finalBalance, _ := decimal.NewFromString(acc.Balance)
		// Should be 1000 + 100 + 200 = 1300 if both succeeded
		// Or 1000 + 100 or 1000 + 200 if one succeeded
		assert.True(t, finalBalance.GreaterThanOrEqual(decimal.NewFromInt(1100)))
	})
}

func TestStore_ExecTx_Atomicity(t *testing.T) {
	t.Run("all operations succeed or all fail", func(t *testing.T) {
		ctx := context.Background()
		account1 := setupTestAccount(t)
		defer cleanupTestAccount(t, account1)

		account2 := setupTestAccount(t)
		defer cleanupTestAccount(t, account2)

		// Set balances
		err := testStore.Queries.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: "500.0000",
			ID:      account1,
		})
		require.NoError(t, err)

		// Try to transfer with an error in the middle
		txID := uuid.New()
		err = testStore.ExecTx(ctx, func(q *sqlc.Queries) error {
			// Debit account1
			_, err := q.CreateEntry(ctx, sqlc.CreateEntryParams{
				AccountID:     account1,
				Debit:         "100.0000",
				Credit:        decimal.Zero.StringFixed(4),
				TransactionID: txID,
				OperationType: "transfer",
				Description:   sql.NullString{String: "Test transfer", Valid: true},
			})
			if err != nil {
				return err
			}

			err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
				Balance: "-100.0000",
				ID:      account1,
			})
			if err != nil {
				return err
			}

			// Simulate error before crediting account2
			return fmt.Errorf("error before credit")
		})

		require.Error(t, err)

		// Verify account1 balance unchanged (rollback worked)
		acc1, err := testStore.Queries.GetAccount(ctx, account1)
		require.NoError(t, err)
		assert.Equal(t, "500.0000", acc1.Balance)

		// Verify no entries created (rollback worked)
		entries, err := testStore.Queries.ListEntriesByTransaction(ctx, txID)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

func TestStore_GetSettlementAccount(t *testing.T) {
	t.Run("gets settlement account", func(t *testing.T) {
		ctx := context.Background()

		settlement, err := testStore.Queries.GetSettlementAccount(ctx)
		require.NoError(t, err)

		assert.True(t, settlement.IsSystem)
		assert.Equal(t, "Settlement Account", settlement.Name)
		assert.Equal(t, "NGN", settlement.Currency)
	})

	t.Run("settlement account locked for update", func(t *testing.T) {
		ctx := context.Background()

		err := testStore.ExecTx(ctx, func(q *sqlc.Queries) error {
			settlement, err := q.GetSettlementAccountForUpdate(ctx)
			if err != nil {
				return err
			}

			assert.True(t, settlement.IsSystem)
			return nil
		})

		require.NoError(t, err)
	})
}

func TestStore_NewStore(t *testing.T) {
	t.Run("creates new store instance", func(t *testing.T) {
		store := NewStore(testDB)
		assert.NotNil(t, store)
		assert.NotNil(t, store.Queries)
	})
}
