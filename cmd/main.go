// Package main wires together the HTTP server, database store, and middleware.
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/PaulBabatuyi/Double-Entry-Bank-Go/docs"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/api"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/db"
	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

// @title           Double-Entry Bank Ledger API
// @version         1.0
// @description     Production-grade double-entry accounting ledger
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

// noDirFileSystem wraps http.FileSystem to disable directory listings.
// Directories that contain an index.html are still served normally.
type noDirFileSystem struct {
	base http.FileSystem
}

func (fs noDirFileSystem) Open(name string) (http.File, error) {
	f, err := fs.base.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	if stat.IsDir() {
		// Only permit directories that have an index.html (e.g. the root).
		idx, idxErr := fs.base.Open(name + "/index.html")
		if idxErr != nil {
			_ = f.Close()
			return nil, os.ErrNotExist
		}
		_ = idx.Close()
	}

	return f, nil
}

func parseAllowedOrigins() []string {
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(origins) == "" {
		return []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	}

	parts := strings.Split(origins, ",")
	allowed := make([]string, 0, len(parts))
	for _, origin := range parts {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed = append(allowed, trimmed)
		}
	}

	if len(allowed) == 0 {
		return []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	}

	return allowed
}

func resolveDBURL() string {
	connStr := strings.TrimSpace(os.Getenv("DB_URL"))

	fallbackVars := []string{"INTERNAL_DATABASE_URL", "RENDER_DATABASE_URL", "DATABASE_URL"}

	if connStr == "" {
		for _, envVar := range fallbackVars {
			if value := strings.TrimSpace(os.Getenv(envVar)); value != "" {
				return value
			}
		}

		// Default connection string for local development only.
		return "postgresql://root:secret@localhost:5432/simple_ledger?sslmode=disable" // #nosec G101 - Local development default
	}

	lower := strings.ToLower(connStr)
	isLocalHostURL := strings.Contains(lower, "@localhost:") || strings.Contains(lower, "@127.0.0.1:") || strings.Contains(lower, "@[::1]:")
	if isLocalHostURL {
		for _, envVar := range fallbackVars {
			if value := strings.TrimSpace(os.Getenv(envVar)); value != "" {
				return value
			}
		}
	}

	return connStr
}

func main() {
	startTime := time.Now()

	initLogger()

	if err := godotenv.Load(); err != nil {
		zlog.Warn().Err(err).Msg("No .env file found – using system env")
	}

	if err := api.InitTokenAuthFromEnv(); err != nil {
		zlog.Fatal().Err(err).Msg("Failed to initialize JWT auth")
	}

	connStr := resolveDBURL()
	if strings.Contains(connStr, "@localhost:") || strings.Contains(connStr, "@127.0.0.1:") || strings.Contains(connStr, "@[::1]:") {
		zlog.Warn().Msg("Using localhost DB_URL; this is only valid for local development")
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

	// CORS middleware for separate frontend deployments and local development.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   parseAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())
			zlog.Info().Str("request_id", reqID).Str("path", r.URL.Path).Msg("Request received")
			next.ServeHTTP(w, r)
		})
	})

	// Serve static frontend files without exposing directory listings
	fileServer := http.FileServer(noDirFileSystem{http.Dir("./frontend")})
	r.Handle("/*", fileServer)

	// Public routes
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		zlog.Info().Msg("Health check requested")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"version": "0.1.0",
			"uptime":  time.Since(startTime).String(),
		}); err != nil {
			zlog.Error().Err(err).Msg("Failed to encode health check response")
		}
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
		r.Get("/accounts/{id}/reconcile", h.ReconcileAccount)
		r.Get("/transactions/{id}", h.GetTransactions)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure HTTP server with timeouts for security
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	zlog.Info().Str("port", port).Msg("Starting server")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zlog.Fatal().Err(err).Msg("Server failed to start")
	}
}
