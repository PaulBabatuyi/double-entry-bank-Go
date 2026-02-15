package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	_ "github.com/PaulBabatuyi/Double-Entry-Bank-Go/docs"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/api"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
)

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Caller().Logger()
	zlog.Info().Msg("Logger initialized")
}

func main() {
	initLogger()

	if err := godotenv.Load(); err != nil {
		zlog.Warn().Err(err).Msg("No .env file found – using system env")
	}

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "postgresql://root:secret@localhost:5432/simple_ledger?sslmode=disable"
		zlog.Warn().Msg("Using default DB_URL – set DB_URL in .env")
	}
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to open DB connection")
	}
	defer dbConn.Close()

	store := db.NewStore(dbConn)
	ledgerSvc := service.NewLedgerService(store)

	h := api.NewHandler(ledgerSvc, store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Public routes
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		zlog.Info().Msg("Health check requested")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"version": "0.1.0",
			"uptime":  time.Since(time.Now()).String(),
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
	))
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(api.TokenAuth))
		r.Use(jwtauth.Authenticator(api.TokenAuth))

		r.Post("/accounts", h.CreateAccount)
		r.Get("/accounts", h.ListAccounts)
		r.Get("/accounts/{id}", h.GetAccount)
		r.Post("/accounts/{id}/deposit", h.Deposit)
		r.Post("/accounts/{id}/withdraw", h.Withdraw)
		r.Post("/transfers", h.Transfer)
		r.Get("/accounts/{id}/entries", h.GetEntries)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	zlog.Info().Str("port", port).Msg("Starting server")
	http.ListenAndServe(":"+port, r)
}
