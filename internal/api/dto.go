package api

import "time"

// AccountResponse represents an account returned by the API.
type AccountResponse struct {
	ID        string    `json:"id"`
	OwnerID   *string   `json:"owner_id,omitempty"`
	Name      string    `json:"name"`
	Balance   string    `json:"balance"`
	Currency  string    `json:"currency"`
	IsSystem  bool      `json:"is_system"`
	CreatedAt time.Time `json:"created_at"`
}

// EntryResponse represents a ledger entry returned by the API.
type EntryResponse struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"account_id"`
	Debit         string    `json:"debit"`
	Credit        string    `json:"credit"`
	TransactionID string    `json:"transaction_id"`
	OperationType string    `json:"operation_type"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// RegisterResponse is returned after successful registration.
type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// TokenResponse contains a signed JWT.
type TokenResponse struct {
	Token string `json:"token"`
}

// MessageResponse contains a simple status message.
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse contains an API error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ReconcileResponse reports whether stored and computed balances match.
type ReconcileResponse struct {
	Matched bool   `json:"matched"`
	Message string `json:"message"`
}
