package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/google/uuid"
)

var (
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrSameAccountTransfer = errors.New("cannot transfer to the same account")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrAccountNotFound     = errors.New("account not found")
)

type LedgerService struct {
	store *db.Store
}

func NewLedgerService(store *db.Store) *LedgerService {
	return &LedgerService{store: store}
}

// Deposit external money into user account
func (s *LedgerService) Deposit(ctx context.Context, accountID uuid.UUID, amountStr string) error {
	if err := validatePositiveAmount(amountStr); err != nil {
		return err
	}

	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		settlement, err := q.GetSettlementAccount(ctx)
		if err != nil {
			return fmt.Errorf("settlement account not found: %w", err)
		}

		account, err := q.GetAccount(ctx, accountID)
		if err != nil {
			return fmt.Errorf("account not found: %w", err)
		}

		if account.Currency != settlement.Currency {
			return ErrCurrencyMismatch
		}

		txID := uuid.New()

		// 1. Credit user account (entry)
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     accountID,
			Debit:         "0.0000",
			Credit:        amountStr,
			TransactionID: txID,
			OperationType: "deposit",
			Description:   sql.NullString{String: "External deposit", Valid: true},
		})
		if err != nil {
			return err
		}

		// 2. Debit settlement (opposing entry)
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     settlement.ID,
			Debit:         amountStr,
			Credit:        "0.0000",
			TransactionID: txID,
			OperationType: "deposit",
			Description:   sql.NullString{String: fmt.Sprintf("Deposit to account %s", accountID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Update balances
		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amountStr,
			ID:      accountID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: fmt.Sprintf("-%s", amountStr),
			ID:      settlement.ID,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

// Withdraw money from user account to external
func (s *LedgerService) Withdraw(ctx context.Context, accountID uuid.UUID, amountStr string) error {
	if err := validatePositiveAmount(amountStr); err != nil {
		return err
	}

	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		settlement, err := q.GetSettlementAccount(ctx)
		if err != nil {
			return err
		}

		account, err := q.GetAccount(ctx, accountID)
		if err != nil {
			return err
		}

		if account.Currency != settlement.Currency {
			return ErrCurrencyMismatch
		}

		// Check sufficient balance (a simple read — serializable tx prevents races)
		// string comparison is safe for same precision/format
		if account.Balance < amountStr {
			return ErrInsufficientFunds
		}

		txID := uuid.New()

		// 1. Debit user
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     accountID,
			Debit:         amountStr,
			Credit:        "0.0000",
			TransactionID: txID,
			OperationType: "withdrawal",
			Description:   sql.NullString{String: "External withdrawal", Valid: true},
		})
		if err != nil {
			return err
		}

		// 2. Credit settlement
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     settlement.ID,
			Debit:         "0.0000",
			Credit:        amountStr,
			TransactionID: txID,
			OperationType: "withdrawal",
			Description:   sql.NullString{String: fmt.Sprintf("Withdrawal from %s", accountID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Update balances
		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: fmt.Sprintf("-%s", amountStr),
			ID:      accountID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amountStr,
			ID:      settlement.ID,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

// Transfer between two user accounts
func (s *LedgerService) Transfer(ctx context.Context, fromID, toID uuid.UUID, amountStr string) error {
	if err := validatePositiveAmount(amountStr); err != nil {
		return err
	}

	if fromID == toID {
		return ErrSameAccountTransfer
	}

	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		fromAcc, err := q.GetAccount(ctx, fromID)
		if err != nil {
			return err
		}

		toAcc, err := q.GetAccount(ctx, toID)
		if err != nil {
			return err
		}

		if fromAcc.Currency != toAcc.Currency {
			return ErrCurrencyMismatch
		}

		if fromAcc.Balance < amountStr {
			return ErrInsufficientFunds
		}

		txID := uuid.New()

		// 1. Debit from account
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     fromID,
			Debit:         amountStr,
			Credit:        "0.0000",
			TransactionID: txID,
			OperationType: "transfer",
			Description:   sql.NullString{String: fmt.Sprintf("Transfer to %s", toID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 2. Credit to account
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     toID,
			Debit:         "0.0000",
			Credit:        amountStr,
			TransactionID: txID,
			OperationType: "transfer",
			Description:   sql.NullString{String: fmt.Sprintf("Transfer from %s", fromID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Update balances
		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: fmt.Sprintf("-%s", amountStr),
			ID:      fromID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amountStr,
			ID:      toID,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

// validate amount string
func validatePositiveAmount(amount string) error {
	// In real project, i will use decimal.Decimal.Parse(amount) > 0, but For now basic string check
	if amount == "" || amount == "0.0000" || amount[0] == '-' {
		return ErrInvalidAmount
	}
	return nil
}
