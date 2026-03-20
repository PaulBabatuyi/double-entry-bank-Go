package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// Dummy sqlc.Queries for demonstration. Replace with real implementation or mocks.
type dummyQueries struct{}

func TestIsSerializationError(t *testing.T) {
	pqErr := &pq.Error{Code: "40001"}
	assert.True(t, isSerializationError(pqErr))
	assert.False(t, isSerializationError(errors.New("some other error")))
}

func TestRetryWait(t *testing.T) {
	assert.Equal(t, 50*time.Millisecond, retryWait(0))
	assert.Equal(t, 100*time.Millisecond, retryWait(1))
	assert.Equal(t, 200*time.Millisecond, retryWait(2))
	assert.Equal(t, time.Second, retryWait(5))
}

func TestSleepWithContext_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sleepWithContext(ctx, 50*time.Millisecond)
	assert.Error(t, err)
}

func TestNewStore(t *testing.T) {
	db := &sql.DB{} // Not a real connection, just for constructor test
	store := NewStore(db)
	assert.NotNil(t, store)
}
