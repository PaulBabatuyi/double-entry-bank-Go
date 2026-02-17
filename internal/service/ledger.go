package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
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
	amount, err := validatePositiveAmount(amountStr)
	if err != nil {
		return err
	}

	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		settlement, err := q.GetSettlementAccountForUpdate(ctx)
		if err != nil {
			return fmt.Errorf("settlement account not found: %w", err)
		}

		account, err := q.GetAccountForUpdate(ctx, accountID)
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
			Debit:         decimal.Zero.StringFixed(4),
			Credit:        amount.StringFixed(4),
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
			Debit:         amount.StringFixed(4),
			Credit:        decimal.Zero.StringFixed(4),
			TransactionID: txID,
			OperationType: "deposit",
			Description:   sql.NullString{String: fmt.Sprintf("Deposit to account %s", accountID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Update balances
		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amount.StringFixed(4),
			ID:      accountID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amount.Neg().StringFixed(4),
			ID:      settlement.ID,
		})
		if err != nil {
			return err
		}

		log.Info().
			Str("tx_id", txID.String()).
			Str("account_id", accountID.String()).
			Str("amount", amount.StringFixed(4)).
			Msg("Deposit completed")

		return nil
	})
}

// Withdraw external money from user account
func (s *LedgerService) Withdraw(ctx context.Context, accountID uuid.UUID, amountStr string) error {
	amount, err := validatePositiveAmount(amountStr)
	if err != nil {
		return err
	}

	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		settlement, err := q.GetSettlementAccountForUpdate(ctx)
		if err != nil {
			return fmt.Errorf("settlement account not found: %w", err)
		}

		account, err := q.GetAccountForUpdate(ctx, accountID)
		if err != nil {
			return fmt.Errorf("account not found: %w", err)
		}

		if account.Currency != settlement.Currency {
			return ErrCurrencyMismatch
		}

		balanceDec, err := decimal.NewFromString(account.Balance)
		if err != nil {
			return errors.New("invalid balance")
		}

		if balanceDec.LessThan(amount) {
			return ErrInsufficientFunds
		}

		txID := uuid.New()

		// 1. Debit user
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     accountID,
			Debit:         amount.StringFixed(4),
			Credit:        decimal.Zero.StringFixed(4),
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
			Debit:         decimal.Zero.StringFixed(4),
			Credit:        amount.StringFixed(4),
			TransactionID: txID,
			OperationType: "withdrawal",
			Description:   sql.NullString{String: fmt.Sprintf("Withdrawal from %s", accountID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Update balances
		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amount.Neg().StringFixed(4),
			ID:      accountID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amount.StringFixed(4),
			ID:      settlement.ID,
		})
		if err != nil {
			return err
		}

		log.Info().
			Str("tx_id", txID.String()).
			Str("account_id", accountID.String()).
			Str("amount", amount.StringFixed(4)).
			Msg("Withdrawal completed")

		return nil
	})
}

// Transfer between two user accounts
func (s *LedgerService) Transfer(ctx context.Context, fromID, toID uuid.UUID, amountStr string) error {
	amount, err := validatePositiveAmount(amountStr)
	if err != nil {
		return err
	}

	if fromID == toID {
		return ErrSameAccountTransfer
	}

	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		fromAcc, err := q.GetAccountForUpdate(ctx, fromID)
		if err != nil {
			return err
		}

		toAcc, err := q.GetAccountForUpdate(ctx, toID)
		if err != nil {
			return err
		}

		if fromAcc.Currency != toAcc.Currency {
			return ErrCurrencyMismatch
		}

		fromBalance, err := decimal.NewFromString(fromAcc.Balance)
		if err != nil {
			return errors.New("invalid from balance")
		}

		if fromBalance.LessThan(amount) {
			return ErrInsufficientFunds
		}

		txID := uuid.New()

		// 1. Debit from
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     fromID,
			Debit:         amount.StringFixed(4),
			Credit:        decimal.Zero.StringFixed(4),
			TransactionID: txID,
			OperationType: "transfer",
			Description:   sql.NullString{String: fmt.Sprintf("Transfer to %s", toID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 2. Credit to
		_, err = q.CreateEntry(ctx, sqlc.CreateEntryParams{
			AccountID:     toID,
			Debit:         decimal.Zero.StringFixed(4),
			Credit:        amount.StringFixed(4),
			TransactionID: txID,
			OperationType: "transfer",
			Description:   sql.NullString{String: fmt.Sprintf("Transfer from %s", fromID), Valid: true},
		})
		if err != nil {
			return err
		}

		// 3. Update balances
		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amount.Neg().StringFixed(4),
			ID:      fromID,
		})
		if err != nil {
			return err
		}

		err = q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
			Balance: amount.StringFixed(4),
			ID:      toID,
		})
		if err != nil {
			return err
		}

		log.Info().
			Str("tx_id", txID.String()).
			Str("from_id", fromID.String()).
			Str("to_id", toID.String()).
			Str("amount", amount.StringFixed(4)).
			Msg("Transfer completed")

		return nil
	})
}

// ReconcileAccount verifies stored balance == SUM(credits) - SUM(debits)
func (s *LedgerService) ReconcileAccount(ctx context.Context, accountID uuid.UUID) (bool, error) {
	account, err := s.store.Queries.GetAccount(ctx, accountID)
	if err != nil {
		return false, fmt.Errorf("account not found: %w", err)
	}

	// Use GetCalculatedBalance which returns interface{} containing the NUMERIC value
	calculatedRaw, err := s.store.Queries.GetCalculatedBalance(ctx, accountID)
	if err != nil {
		return false, fmt.Errorf("failed to calculate balance: %w", err)
	}

	var calculatedStr string
	switch v := calculatedRaw.(type) {
	case []byte:
		calculatedStr = string(v)
	case string:
		calculatedStr = v
	case int64:
		calculatedStr = fmt.Sprintf("%d.0000", v)
	case float64:
		calculatedStr = fmt.Sprintf("%.4f", v)
	default:
		return false, fmt.Errorf("unexpected calculated balance type: %T", calculatedRaw)
	}

	calculated, err := decimal.NewFromString(calculatedStr)
	if err != nil {
		return false, fmt.Errorf("invalid calculated balance: %w", err)
	}

	stored, err := decimal.NewFromString(account.Balance)
	if err != nil {
		return false, fmt.Errorf("invalid stored balance: %w", err)
	}

	if !stored.Equal(calculated) {
		log.Error().
			Str("account_id", accountID.String()).
			Str("stored_balance", account.Balance).
			Str("calculated", calculated.StringFixed(4)).
			Msg("Balance mismatch detected")
		return false, fmt.Errorf("balance mismatch: stored %s, calculated %s",
			account.Balance, calculated.StringFixed(4))
	}

	log.Info().
		Str("account_id", accountID.String()).
		Str("balance", account.Balance).
		Msg("Account reconciled successfully")

	return true, nil
}

// validatePositiveAmount parses and validates that amount > 0
func validatePositiveAmount(amountStr string) (decimal.Decimal, error) {
	amt, err := decimal.NewFromString(amountStr)
	if err != nil {
		return decimal.Zero, ErrInvalidAmount
	}
	if amt.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrInvalidAmount
	}
	return amt, nil
}
