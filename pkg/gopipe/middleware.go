package gopipe

import (
	"context"
	"fmt"
	"gopipe/internal/utils"
	"math/rand"
	"time"
)

// Middleware defines processing middleware
type Middleware func(Handler) Handler

// Handler processes a task
type Handler func(context.Context, *Task) error

// Chain middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(final Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// LoggingMiddleware logs task execution
func LoggingMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, task *Task) error {
			start := time.Now()
			fmt.Printf("Starting task %s at %s\n", task.ID, start.Format(time.RFC3339))

			err := next(ctx, task)

			duration := time.Since(start)
			if err != nil {
				fmt.Printf("Task %s failed after %v: %v\n", task.ID, duration, err)
			} else {
				fmt.Printf("Task %s completed in %v\n", task.ID, duration)
			}

			return err
		}
	}
}

// RetryMiddleware handles retries with backoff
func RetryMiddleware(cfg utils.BackoffConfig) Middleware {
	backoff := utils.NewBackoff(cfg)

	return func(next Handler) Handler {
		return func(ctx context.Context, task *Task) error {
			var lastErr error
			for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
				if attempt > 0 {
					delay := backoff.Calculate(attempt)
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				if err := next(ctx, task); err == nil {
					return nil
				} else {
					lastErr = err
					fmt.Printf("Task %s failed on attempt %d: %v\n", task.ID, attempt+1, err)
				}
			}
			return fmt.Errorf("max retries reached for task %s: %w", task.ID, lastErr)
		}
	}
}

// CircuitBreakerMiddleware implements circuit breaker pattern
func CircuitBreakerMiddleware(cb *CircuitBreaker) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, task *Task) error {
			if !cb.Allow() {
				return ErrCircuitOpen
			}

			err := next(ctx, task)
			if err != nil {
				cb.RecordFailure()
			} else {
				cb.RecordSuccess()
			}

			return err
		}
	}
}

// TimeoutMiddleware adds per-handler timeout
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, task *Task) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			return next(ctx, task)
		}
	}
}

// BackoffFunc calculates backoff duration
type BackoffFunc func(attempt int) time.Duration

// ExponentialBackoff with jitter
func ExponentialBackoff(base time.Duration, max time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		if attempt == 0 {
			return 0
		}

		backoff := base * time.Duration(1<<uint(attempt-1))
		if backoff > max {
			backoff = max
		}

		// Add jitter
		jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
		return backoff/2 + jitter
	}
}

func isRetryableError(err error) bool {
	// Define which errors are retryable
	return err != ErrHandlerNotFound && err != context.Canceled
}

func RateLimitMiddleware(rl *utils.RateLimiter) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, task *Task) error {
			if !rl.Allow() {
				return ErrRateLimited
			}
			return next(ctx, task)
		}
	}
}
