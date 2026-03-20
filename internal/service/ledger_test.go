package service

import (
	"context"
	"sync"
	"testing"

	"database/sql"
	"os"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	_ "github.com/lib/pq"
)

// setupTestLedger and helpers would be implemented to provide a testable LedgerService and test DB.
// For demonstration, these are placeholders. In a real repo, use test containers or a test DB.

func setupTestLedger(t *testing.T) *LedgerService {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://root:secret@localhost:5432/simple_ledger?sslmode=disable"
	}
	sqlDB, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	store := db.NewStore(sqlDB)
	ledger := NewLedgerService(store)
	return ledger
}

func createTestAccount(t *testing.T, ledger *LedgerService, balance string) uuid.UUID {
	// Use a unique account name for each test run
	accName := "Test Account " + uuid.New().String()
	
	// Get settlement account to match its currency
	settlement, err := ledger.store.Queries.GetSettlementAccount(context.Background())
	require.NoError(t, err)
	
	account, err := ledger.store.Queries.CreateAccount(context.Background(), sqlc.CreateAccountParams{
		OwnerID:  uuid.NullUUID{Valid: false}, // No owner for test accounts
		Name:     accName,
		Currency: settlement.Currency, // Match settlement account currency
		IsSystem: false,
	})
	require.NoError(t, err)
	// Fund the account if needed
	if balance != "0.00" && balance != "0" && balance != "" {
		err = ledger.Deposit(context.Background(), account.ID, balance)
		require.NoError(t, err)
	}
	return account.ID
}

func getAccountBalance(t *testing.T, ledger *LedgerService, accountID uuid.UUID) string {
	balance, err := ledger.store.Queries.GetAccountBalance(context.Background(), accountID)
	require.NoError(t, err)
	return balance
}

func TestDeposit_Success(t *testing.T) {
	ledger := setupTestLedger(t)
	accountID := createTestAccount(t, ledger, "0.00")
	err := ledger.Deposit(context.Background(), accountID, "100.00")
	require.NoError(t, err)
	balance := getAccountBalance(t, ledger, accountID)
	assert.Equal(t, "100.0000", balance)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	ledger := setupTestLedger(t)
	accountID := createTestAccount(t, ledger, "50.00")
	err := ledger.Withdraw(context.Background(), accountID, "100.00")
	assert.Error(t, err)
	// Optionally check for ErrInsufficientFunds
}

func TestConcurrentDeposits(t *testing.T) {
	ledger := setupTestLedger(t)
	accountID := createTestAccount(t, ledger, "0.00")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = ledger.Deposit(context.Background(), accountID, "100.00")
	}()
	go func() {
		defer wg.Done()
		_ = ledger.Deposit(context.Background(), accountID, "100.00")
	}()
	wg.Wait()
	balance := getAccountBalance(t, ledger, accountID)
	assert.Equal(t, "200.0000", balance)
}
