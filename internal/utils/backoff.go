package utils

import (
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy defines backoff strategies
type BackoffStrategy int

const (
	StrategyExponential BackoffStrategy = iota
	StrategyLinear
	StrategyConstant
	StrategyFibonacci
)

// BackoffConfig configuration for backoff
type BackoffConfig struct {
	Strategy    BackoffStrategy
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	MaxAttempts int
	Jitter      bool
}

// Backoff calculates backoff duration for retry attempts
type Backoff struct {
	config BackoffConfig
}

// NewBackoff creates a new backoff calculator
func NewBackoff(config BackoffConfig) *Backoff {
	return &Backoff{config: config}
}

// Calculate returns the backoff duration for the given attempt
func (b *Backoff) Calculate(attempt int) time.Duration {
	if attempt >= b.config.MaxAttempts {
		return b.config.MaxDelay
	}

	var delay time.Duration
	
	switch b.config.Strategy {
	case StrategyExponential:
		delay = b.exponentialBackoff(attempt)
	case StrategyLinear:
		delay = b.linearBackoff(attempt)
	case StrategyConstant:
		delay = b.config.BaseDelay
	case StrategyFibonacci:
		delay = b.fibonacciBackoff(attempt)
	default:
		delay = b.exponentialBackoff(attempt)
	}

	// Apply jitter if enabled
	if b.config.Jitter {
		delay = b.applyJitter(delay)
	}

	// Cap at max delay
	if delay > b.config.MaxDelay {
		delay = b.config.MaxDelay
	}

	return delay
}

func (b *Backoff) exponentialBackoff(attempt int) time.Duration {
	delay := b.config.BaseDelay * time.Duration(math.Pow(2, float64(attempt)))
	return delay
}

func (b *Backoff) linearBackoff(attempt int) time.Duration {
	return b.config.BaseDelay * time.Duration(attempt+1)
}

func (b *Backoff) fibonacciBackoff(attempt int) time.Duration {
	fib := func(n int) int {
		if n <= 1 {
			return n
		}
		a, b := 0, 1
		for i := 2; i <= n; i++ {
			a, b = b, a+b
		}
		return b
	}
	return b.config.BaseDelay * time.Duration(fib(attempt+1))
}

func (b *Backoff) applyJitter(delay time.Duration) time.Duration {
	jitter := rand.Float64() * 0.3 // ±30% jitter
	jittered := float64(delay) * (1 + jitter - 0.15)
	return time.Duration(jittered)
}

// DefaultBackoff returns a sensible default backoff configuration
func DefaultBackoff() BackoffConfig {
	return BackoffConfig{
		Strategy:    StrategyExponential,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    30 * time.Second,
		MaxAttempts: 5,
		Jitter:      true,
	}
}