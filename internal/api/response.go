package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	// Centralize JSON response writing so handlers stay focused on business flow.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	// Keep API error shape consistent across every endpoint.
	respondJSON(w, status, ErrorResponse{Error: msg})
}
