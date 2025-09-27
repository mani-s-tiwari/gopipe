package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
)

func main() {
	// Create pipeline
	pipeline := gopipe.NewPipeline()

	// Create circuit breaker for image processing
	cb := gopipe.NewCircuitBreaker(5, 30*time.Second)

	// Register handlers with middleware chains
	pipeline.RegisterHandler(
		"resize_image",
		resizeImageHandler,
		gopipe.LoggingMiddleware(),
		gopipe.CircuitBreakerMiddleware(cb),
		gopipe.RetryMiddleware(gopipe.DefaultBackoff()), // now uses utils.BackoffConfig
		gopipe.TimeoutMiddleware(10*time.Second),
	)

	pipeline.RegisterHandler(
		"send_email",
		sendEmailHandler,
		gopipe.LoggingMiddleware(),
		gopipe.RetryMiddleware(gopipe.DefaultBackoff()),
	)

	// ---- Submit async tasks ----
	task1 := gopipe.NewTask(
		"resize_image",
		[]byte(`{"file": "image.jpg", "width": 800}`),
		gopipe.WithPriority(gopipe.PriorityHigh),
		gopipe.WithMetadata("user_id", "12345"),
	)

	task2 := gopipe.NewTask(
		"send_email",
		[]byte(`{"to": "user@example.com", "subject": "Welcome"}`),
		gopipe.WithScheduledTime(time.Now().Add(5*time.Minute)), // schedule in future
	)

	pipeline.Submit(task1)
	pipeline.Submit(task2)

	// ---- Submit and wait for result ----
	task3 := gopipe.NewTask(
		"resize_image",
		[]byte(`{"file": "critical.jpg"}`),
		gopipe.WithPriority(gopipe.PriorityCritical),
	)

	result := pipeline.SubmitAndWait(task3)
	fmt.Printf("[main] Task completed: %+v\n", result)

	// ---- Metrics monitoring ----
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := pipeline.GetMetrics()
			fmt.Printf("[metrics] submitted=%d completed=%d failed=%d avgLatency=%v\n",
				metrics.TasksSubmitted.Get(),
				metrics.TasksCompleted.Get(),
				metrics.TasksFailed.Get(),
				metrics.TaskLatency.Avg(),
			)
		}
	}()

	// Run for demo then graceful shutdown
	time.Sleep(1 * time.Minute)
	pipeline.Stop()
}

// resizeImageHandler simulates image processing
func resizeImageHandler(ctx context.Context, task *gopipe.Task) error {
	fmt.Printf("[handler] resizing image: %s\n", string(task.Payload))
	select {
	case <-time.After(2 * time.Second): // simulate work
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// sendEmailHandler simulates sending an email
func sendEmailHandler(ctx context.Context, task *gopipe.Task) error {
	fmt.Printf("[handler] sending email: %s\n", string(task.Payload))
	select {
	case <-time.After(1 * time.Second): // simulate work
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
