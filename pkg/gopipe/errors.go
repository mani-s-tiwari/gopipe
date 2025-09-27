package gopipe

import (
	"errors"
	"fmt"
	"time"
)

// Error types for precise handling
var (
	ErrHandlerNotFound = errors.New("handler not found")
	ErrQueueFull       = errors.New("queue is full")
	ErrTimeout         = errors.New("operation timed out")
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrRateLimited     = errors.New("rate limit exceeded")
)

// ErrorWithRetry includes retry information
type ErrorWithRetry struct {
	Err       error
	Retryable bool
	RetryAfter time.Duration
	Operation string
}

func (e ErrorWithRetry) Error() string {
	return fmt.Sprintf("%s: %v (retryable: %v)", e.Operation, e.Err, e.Retryable)
}

// CircuitBreaker tracks failures and can open circuit
type CircuitBreaker struct {
	failures     int
	threshold    int
	timeout      time.Duration
	lastFailure  time.Time
	state        string // closed, open, half-open
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
		state:     "closed",
	}
}

func (cb *CircuitBreaker) Allow() bool {
	if cb.state == "open" && time.Since(cb.lastFailure) < cb.timeout {
		return false
	}
	if cb.state == "open" {
		cb.state = "half-open"
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.failures = 0
	cb.state = "closed"
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = "open"
		cb.lastFailure = time.Now()
	}
}