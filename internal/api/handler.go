package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/service"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type Handler struct {
	ledger *service.LedgerService
	store  *db.Store
}

func NewHandler(ledger *service.LedgerService, store *db.Store) *Handler {
	return &Handler{ledger: ledger, store: store}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
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
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	if input.Email == "" || input.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password required")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.store.Queries.CreateUser(r.Context(), sqlc.CreateUserParams{
		Email:          input.Email,
		HashedPassword: string(hashed),
	})
	if err != nil {
		respondError(w, http.StatusConflict, "user already exists or failed")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"token":   token,
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
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	user, err := h.store.Queries.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// CreateAccount godoc
// @Summary      Create a new account
// @Description  Creates a new user-owned account with name and currency
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        body    body      object{name=string}  true  "Account details"
// @Success      201     {object}  sqlc.Account
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts [post]
// @Security     Bearer
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		respondError(w, http.StatusBadRequest, "name required")
		return
	}

	acc, err := h.store.Queries.CreateAccount(r.Context(), sqlc.CreateAccountParams{
		OwnerID:  uuid.NullUUID{UUID: userID, Valid: true},
		Name:     input.Name,
		Currency: "NGN",
		IsSystem: false,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	respondJSON(w, http.StatusCreated, acc)
}

// ListAccounts godoc
// @Summary      List user accounts
// @Description  Returns list of accounts owned by authenticated user
// @Tags         accounts
// @Produce      json
// @Success      200     {array}   sqlc.Account
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts [get]
// @Security     Bearer
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userIDStr, _ := claims["user_id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	accounts, err := h.store.Queries.ListAccountsByOwner(r.Context(), uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	respondJSON(w, http.StatusOK, accounts)
}

// GetAccount godoc
// @Summary      Get account details
// @Description  Returns details of a specific account
// @Tags         accounts
// @Produce      json
// @Param        id   path      string  true  "Account ID"
// @Success      200  {object}  sqlc.Account
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /accounts/{id} [get]
// @Security     Bearer
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.store.Queries.GetAccount(r.Context(), accountID)
	if err != nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	respondJSON(w, http.StatusOK, acc)
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
// @Router       /accounts/{id}/deposit [post]
// @Security     Bearer
func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var input struct {
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	err = h.ledger.Deposit(r.Context(), accountID, input.Amount)
	if err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, service.ErrInsufficientFunds) || errors.Is(err, service.ErrCurrencyMismatch) {
			code = http.StatusBadRequest
		} else {
			code = http.StatusInternalServerError
		}
		respondError(w, code, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "deposit successful"})
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
// @Router       /accounts/{id}/withdraw [post]
// @Security     Bearer
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var input struct {
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	err = h.ledger.Withdraw(r.Context(), accountID, input.Amount)
	if err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, service.ErrInsufficientFunds) {
			code = http.StatusBadRequest
		}
		respondError(w, code, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "withdrawal successful"})
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
// @Router       /transfers [post]
// @Security     Bearer
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FromID uuid.UUID `json:"from_id"`
		ToID   uuid.UUID `json:"to_id"`
		Amount string    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	err := h.ledger.Transfer(r.Context(), input.FromID, input.ToID, input.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "transfer successful"})
}

// GetEntries godoc
// @Summary      Get account entries
// @Description  Returns list of ledger entries for an account (immutable history)
// @Tags         accounts
// @Produce      json
// @Param        id      path      string  true   "Account ID"
// @Param        limit   query     int     false  "Limit (default 20)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Success      200     {array}   sqlc.Entry
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts/{id}/entries [get]
// @Security     Bearer
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	limit := 20
	offset := 0

	entries, err := h.store.Queries.ListEntriesByAccount(r.Context(), sqlc.ListEntriesByAccountParams{
		AccountID: accountID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch entries")
		return
	}

	respondJSON(w, http.StatusOK, entries)
}
