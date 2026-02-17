package api

import (
	"time"

	"github.com/google/uuid"
)

type AccountResponse struct {
	ID         uuid.UUID  `json:"id"`
	OwnerID    *uuid.UUID `json:"owner_id,omitempty"` // nullable
	Name       string     `json:"name"`
	Balance    string     `json:"balance"`
	Currency   string     `json:"currency"`
	IsSystem   bool       `json:"is_system"`
	CreatedAt  time.Time  `json:"created_at"`
}

type EntriesResponse []EntryResponse

type EntryResponse struct {
	ID            uuid.UUID `json:"id"`
	AccountID     uuid.UUID `json:"account_id"`
	Debit         string    `json:"debit"`
	Credit        string    `json:"credit"`
	TransactionID uuid.UUID `json:"transaction_id"`
	OperationType string    `json:"operation_type"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}


type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// swagger
type TokenResponse struct {
	Token string `json:"token"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
