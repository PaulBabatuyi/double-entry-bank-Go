package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAmountInput(t *testing.T) {
	// String input is accepted as-is after normalization.
	val, err := normalizeAmountInput("100.00")
	assert.NoError(t, err)
	assert.Equal(t, "100.00", val)
}

func TestDecodeAmountFromBody_Invalid(t *testing.T) {
	// Empty body should fail JSON decoding.
	req := &http.Request{Body: http.NoBody}
	_, err := decodeAmountFromBody(req)
	assert.Error(t, err)
}
