// Package api exposes HTTP handlers, middleware, and response types for the ledger service.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/service"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
)

// Handler serves HTTP requests backed by the ledger and store layers.
type Handler struct {
	ledger *service.LedgerService
	store  *db.Store
}

// NewHandler constructs a Handler with the required service and persistence dependencies.
func NewHandler(ledger *service.LedgerService, store *db.Store) *Handler {
	return &Handler{ledger: ledger, store: store}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user with email and hashed password, returns user details and JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      object{email=string,password=string}  true  "User registration details"
// @Success      201     {object}  RegisterResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      409     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn().Err(err).Msg("Failed to decode register request")
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	if input.Email == "" || input.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password required")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.store.CreateUser(r.Context(), sqlc.CreateUserParams{
		Email:          input.Email,
		HashedPassword: string(hashed),
	})
	if err != nil {
		log.Error().Err(err).Str("email", input.Email).Msg("Failed to create user")
		respondError(w, http.StatusConflict, "user already exists or failed")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to generate token")
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("User registered successfully")
	respondJSON(w, http.StatusCreated, RegisterResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
		Token:  token,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticates user with email/password and returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      object{email=string,password=string}  true  "User login details"
// @Success      200     {object}  TokenResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn().Err(err).Msg("Failed to decode login request")
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		log.Warn().Err(err).Str("email", input.Email).Msg("Login failed - user not found")
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); compareErr != nil {
		log.Warn().Str("email", input.Email).Msg("Login failed - invalid password")
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to generate token")
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("User logged in successfully")
	respondJSON(w, http.StatusOK, TokenResponse{Token: token})
}

// CreateAccount godoc
// @Summary      Create a new account
// @Description  Creates a new user-owned account with name and currency
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        body    body      object{name=string}  true  "Account details"
// @Success      201     {object}  AccountResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts [post]
// @Security     Bearer
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if decodeErr := json.NewDecoder(r.Body).Decode(&input); decodeErr != nil || input.Name == "" {
		respondError(w, http.StatusBadRequest, "name required")
		return
	}

	acc, err := h.store.CreateAccount(r.Context(), sqlc.CreateAccountParams{
		OwnerID:  uuid.NullUUID{UUID: userID, Valid: true},
		Name:     input.Name,
		Currency: "NGN",
		IsSystem: false,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("name", input.Name).Msg("Failed to create account")
		respondError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	log.Info().Str("account_id", acc.ID.String()).Str("user_id", userID.String()).Str("name", acc.Name).Msg("Account created")
	respondJSON(w, http.StatusCreated, toAccountResponse(acc))
}

// ListAccounts godoc
// @Summary      List user accounts
// @Description  Returns list of accounts owned by authenticated user
// @Tags         accounts
// @Produce      json
// @Success      200     {array}   AccountResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts [get]
// @Security     Bearer
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	accounts, err := h.store.ListAccountsByOwner(r.Context(), uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list accounts")
		respondError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	response := make([]AccountResponse, len(accounts))
	for i, acc := range accounts {
		response[i] = toAccountResponse(acc)
	}

	respondJSON(w, http.StatusOK, response)
}

// GetAccount godoc
// @Summary      Get account details
// @Description  Returns details of a specific account
// @Tags         accounts
// @Produce      json
// @Param        id   path      string  true  "Account ID"
// @Success      200  {object}  AccountResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /accounts/{id} [get]
// @Security     Bearer
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.store.GetAccount(r.Context(), accountID)
	if err != nil {
		log.Warn().Err(err).Str("account_id", accountID.String()).Msg("Account not found")
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	if acc.OwnerID.Valid && acc.OwnerID.UUID != userID {
		log.Warn().Str("account_id", accountID.String()).Str("user_id", userID.String()).Str("owner_id", acc.OwnerID.UUID.String()).Msg("Access denied to account")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	respondJSON(w, http.StatusOK, toAccountResponse(acc))
}

// Deposit godoc
// @Summary      Deposit money into account
// @Description  Deposits fiat amount (mock) with double-entry ledger update
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id      path      string  true   "Account ID"
// @Param        body    body      object{amount=string}  true  "Deposit amount (e.g., 1000.0000)"
// @Success      200     {object}  MessageResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      403     {object}  ErrorResponse
// @Router       /accounts/{id}/deposit [post]
// @Security     Bearer
func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.store.GetAccount(r.Context(), accountID)
	if err != nil {
		log.Warn().Err(err).Str("account_id", accountID.String()).Msg("Deposit failed - account not found")
		respondError(w, http.StatusNotFound, "account not found")
		return
	}
	if acc.OwnerID.Valid && acc.OwnerID.UUID != userID {
		log.Warn().Str("account_id", accountID.String()).Str("user_id", userID.String()).Msg("Deposit denied - access forbidden")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	amount, err := decodeAmountFromBody(r)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to decode deposit request")
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	err = h.ledger.Deposit(r.Context(), accountID, amount)
	if err != nil {
		log.Error().Err(err).Str("account_id", accountID.String()).Str("amount", amount).Msg("Deposit failed")
		code := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidAmount) || errors.Is(err, service.ErrCurrencyMismatch) {
			code = http.StatusBadRequest
		}
		respondError(w, code, err.Error())
		return
	}

	log.Info().Str("account_id", accountID.String()).Str("user_id", userID.String()).Str("amount", amount).Msg("Deposit successful")
	respondJSON(w, http.StatusOK, MessageResponse{Message: "deposit successful"})
}

// Withdraw godoc
// @Summary      Withdraw money from account
// @Description  Withdraws fiat amount (mock) with double-entry ledger update
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id      path      string  true   "Account ID"
// @Param        body    body      object{amount=string}  true  "Withdraw amount (e.g., 500.0000)"
// @Success      200     {object}  MessageResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      403     {object}  ErrorResponse
// @Router       /accounts/{id}/withdraw [post]
// @Security     Bearer
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.store.GetAccount(r.Context(), accountID)
	if err != nil {
		log.Warn().Err(err).Str("account_id", accountID.String()).Msg("Withdrawal failed - account not found")
		respondError(w, http.StatusNotFound, "account not found")
		return
	}
	if acc.OwnerID.Valid && acc.OwnerID.UUID != userID {
		log.Warn().Str("account_id", accountID.String()).Str("user_id", userID.String()).Msg("Withdrawal denied - access forbidden")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	amount, err := decodeAmountFromBody(r)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to decode withdrawal request")
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	err = h.ledger.Withdraw(r.Context(), accountID, amount)
	if err != nil {
		log.Error().Err(err).Str("account_id", accountID.String()).Str("amount", amount).Msg("Withdrawal failed")
		code := http.StatusInternalServerError
		if errors.Is(err, service.ErrInsufficientFunds) || errors.Is(err, service.ErrInvalidAmount) || errors.Is(err, service.ErrCurrencyMismatch) {
			code = http.StatusBadRequest
		}
		respondError(w, code, err.Error())
		return
	}

	log.Info().Str("account_id", accountID.String()).Str("user_id", userID.String()).Str("amount", amount).Msg("Withdrawal successful")
	respondJSON(w, http.StatusOK, MessageResponse{Message: "withdrawal successful"})
}

// Transfer godoc
// @Summary      Transfer money between accounts
// @Description  Transfers fiat amount (mock) with double-entry ledger update
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        body    body      object{from_id=string,to_id=string,amount=string}  true  "Transfer details"
// @Success      200     {object}  MessageResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      403     {object}  ErrorResponse
// @Router       /transfers [post]
// @Security     Bearer
// Transfer handles money transfers between accounts.
// @Summary Transfer between accounts
// @Description Transfers funds between accounts. The `amount` field accepts either a JSON number or a JSON string. `from_id`/`to_id` are the preferred fields; `from_account_id`/`to_account_id` are supported as legacy aliases.
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	var input struct {
		Amount        interface{} `json:"amount"`
		FromID        string      `json:"from_id"`
		ToID          string      `json:"to_id"`
		FromAccountID string      `json:"from_account_id"`
		ToAccountID   string      `json:"to_account_id"`
	}
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	if decodeErr := dec.Decode(&input); decodeErr != nil {
		log.Warn().Err(decodeErr).Msg("Failed to decode transfer request")
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	fromIDRaw := strings.TrimSpace(input.FromID)
	if fromIDRaw == "" {
		fromIDRaw = strings.TrimSpace(input.FromAccountID)
	}
	toIDRaw := strings.TrimSpace(input.ToID)
	if toIDRaw == "" {
		toIDRaw = strings.TrimSpace(input.ToAccountID)
	}

	log.Info().Str("from_id", fromIDRaw).Str("to_id", toIDRaw).Interface("amount", input.Amount).Msg("Transfer request received")

	if fromIDRaw == "" {
		log.Warn().Msg("Transfer missing from_id")
		respondError(w, http.StatusBadRequest, "from_id (or from_account_id) is required")
		return
	}
	if toIDRaw == "" {
		log.Warn().Msg("Transfer missing to_id")
		respondError(w, http.StatusBadRequest, "to_id (or to_account_id) is required")
		return
	}

	fromID, err := uuid.Parse(fromIDRaw)
	if err != nil {
		log.Warn().Err(err).Str("from_id", fromIDRaw).Msg("Invalid from_id UUID format")
		respondError(w, http.StatusBadRequest, "invalid from_id format")
		return
	}

	toID, err := uuid.Parse(toIDRaw)
	if err != nil {
		log.Warn().Err(err).Str("to_id", toIDRaw).Msg("Invalid to_id UUID format")
		respondError(w, http.StatusBadRequest, "invalid to_id format")
		return
	}

	if fromID == uuid.Nil || toID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "from_id and to_id must be valid non-zero UUIDs")
		return
	}

	amount, err := normalizeAmountInput(input.Amount)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse transfer amount")
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	fromAcc, err := h.store.GetAccount(r.Context(), fromID)
	if err != nil {
		log.Warn().Err(err).Str("from_id", fromID.String()).Msg("Transfer failed - from account not found")
		respondError(w, http.StatusNotFound, "from account not found")
		return
	}
	if fromAcc.OwnerID.Valid && fromAcc.OwnerID.UUID != userID {
		log.Warn().Str("from_id", fromID.String()).Str("user_id", userID.String()).Msg("Transfer denied - access forbidden")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	err = h.ledger.Transfer(r.Context(), fromID, toID, amount)
	if err != nil {
		log.Error().Err(err).Str("from_id", fromID.String()).Str("to_id", toID.String()).Str("amount", amount).Msg("Transfer failed")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info().Str("from_id", fromID.String()).Str("to_id", toID.String()).Str("user_id", userID.String()).Str("amount", amount).Msg("Transfer successful")
	respondJSON(w, http.StatusOK, MessageResponse{Message: "transfer successful"})
}

// GetEntries godoc
// @Summary      Get account entries
// @Description  Returns list of ledger entries for an account (immutable history)
// @Tags         accounts
// @Produce      json
// @Param        id      path      string  true   "Account ID"
// @Param        limit   query     int     false  "Limit (default 20)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Success      200     {array}   EntryResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      403     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts/{id}/entries [get]
// @Security     Bearer
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.store.GetAccount(r.Context(), accountID)
	if err != nil {
		log.Warn().Err(err).Str("account_id", accountID.String()).Msg("Get entries failed - account not found")
		respondError(w, http.StatusNotFound, "account not found")
		return
	}
	if acc.OwnerID.Valid && acc.OwnerID.UUID != userID {
		log.Warn().Str("account_id", accountID.String()).Str("user_id", userID.String()).Msg("Get entries denied - access forbidden")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if v, parseErr := strconv.Atoi(limitStr); parseErr == nil && v > 0 {
		limit = min(v, 100)
	}
	if v, parseErr := strconv.Atoi(offsetStr); parseErr == nil && v >= 0 {
		offset = v
	}

	// Validate conversions to int32 to prevent overflow
	if limit > 2147483647 || offset > 2147483647 {
		respondError(w, http.StatusBadRequest, "limit or offset too large")
		return
	}

	entries, err := h.store.ListEntriesByAccount(r.Context(), sqlc.ListEntriesByAccountParams{
		AccountID: accountID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		log.Error().Err(err).Str("account_id", accountID.String()).Msg("Failed to fetch entries")
		respondError(w, http.StatusInternalServerError, "failed to fetch entries")
		return
	}

	response := make([]EntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = toEntryResponse(entry)
	}

	respondJSON(w, http.StatusOK, response)
}

// GetTransactions godoc
// @Summary      Get transaction details
// @Description  Returns both entries (debit and credit) for a complete transaction view
// @Tags         accounts
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {array}   EntryResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /transactions/{id} [get]
// @Security     Bearer
func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	transactionIDStr := chi.URLParam(r, "id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	entries, err := h.store.ListEntriesByTransaction(r.Context(), transactionID)
	if err != nil {
		log.Error().Err(err).Str("transaction_id", transactionID.String()).Msg("Failed to fetch transaction")
		respondError(w, http.StatusInternalServerError, "failed to fetch transaction")
		return
	}

	if len(entries) == 0 {
		log.Warn().Str("transaction_id", transactionID.String()).Msg("Transaction not found")
		respondError(w, http.StatusNotFound, "transaction not found")
		return
	}

	authorized := false
	for _, entry := range entries {
		acc, err := h.store.GetAccount(r.Context(), entry.AccountID)
		if err != nil {
			log.Error().Err(err).Str("account_id", entry.AccountID.String()).Msg("Failed to authorize transaction")
			respondError(w, http.StatusInternalServerError, "failed to authorize transaction")
			return
		}

		if acc.OwnerID.Valid && acc.OwnerID.UUID == userID {
			authorized = true
			break
		}
	}

	if !authorized {
		log.Warn().Str("transaction_id", transactionID.String()).Str("user_id", userID.String()).Msg("Get transaction denied - access forbidden")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	response := make([]EntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = toEntryResponse(entry)
	}

	respondJSON(w, http.StatusOK, response)
}

// ReconcileAccount godoc
// @Summary      Reconcile account balance
// @Description  Verifies stored balance matches sum of all ledger entries (credits - debits)
// @Tags         accounts
// @Produce      json
// @Param        id   path      string  true  "Account ID"
// @Success      200  {object}  ReconcileResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /accounts/{id}/reconcile [get]
// @Security     Bearer
func (h *Handler) ReconcileAccount(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract JWT from context")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		log.Warn().Msg("user_id claim missing or invalid in JWT")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user_id UUID in token")
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.store.GetAccount(r.Context(), accountID)
	if err != nil {
		log.Warn().Err(err).Str("account_id", accountID.String()).Msg("Reconcile failed - account not found")
		respondError(w, http.StatusNotFound, "account not found")
		return
	}
	if acc.OwnerID.Valid && acc.OwnerID.UUID != userID {
		log.Warn().Str("account_id", accountID.String()).Str("user_id", userID.String()).Msg("Reconcile denied - access forbidden")
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	matched, err := h.ledger.ReconcileAccount(r.Context(), accountID)
	if err != nil {
		log.Error().Err(err).Str("account_id", accountID.String()).Msg("Reconciliation failed")
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Info().Str("account_id", accountID.String()).Bool("matched", matched).Msg("Reconciliation completed")
	respondJSON(w, http.StatusOK, ReconcileResponse{
		Matched: matched,
		Message: "Account reconciled successfully",
	})
}
