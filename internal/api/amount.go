package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func normalizeAmountInput(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		amount := strings.TrimSpace(v)
		if amount == "" {
			return "", errors.New("amount required")
		}
		return amount, nil
	case json.Number:
		amount := strings.TrimSpace(v.String())
		if amount == "" {
			return "", errors.New("amount required")
		}
		return amount, nil
	case float64:
		// Kept as a safety fallback when UseNumber is not enabled.
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", errors.New("amount must be a number or string")
	}
}

func decodeAmountFromBody(r *http.Request) (string, error) {
	var input struct {
		Amount interface{} `json:"amount"`
	}

	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	if err := dec.Decode(&input); err != nil {
		return "", err
	}

	return normalizeAmountInput(input.Amount)
}
