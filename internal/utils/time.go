package utils

import (
	"sync"
	"time"
)

// Timer provides a reusable timer utility
type Timer struct {
	start time.Time
}

// NewTimer creates a new timer
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Start resets and starts the timer
func (t *Timer) Start() *Timer {
	t.start = time.Now()
	return t
}

// Elapsed returns the elapsed time since start
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// ElapsedMs returns elapsed time in milliseconds
func (t *Timer) ElapsedMs() int64 {
	return t.Elapsed().Milliseconds()
}

// Debouncer provides debounce functionality
type Debouncer struct {
	duration time.Duration
	timer    *time.Timer
	mu       sync.Mutex
}

// NewDebouncer creates a new debouncer
func NewDebouncer(duration time.Duration) *Debouncer {
	return &Debouncer{
		duration: duration,
	}
}

// Debounce executes the function after the duration, resetting on subsequent calls
func (d *Debouncer) Debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.duration, f)
}

// Throttler provides throttle functionality
type Throttler struct {
	duration time.Duration
	last     time.Time
	mu       sync.Mutex
}

// NewThrottler creates a new throttler
func NewThrottler(duration time.Duration) *Throttler {
	return &Throttler{
		duration: duration,
		last:     time.Now().Add(-duration), // Allow immediate first call
	}
}

// Throttle executes the function only if the duration has passed since last call
func (t *Throttler) Throttle(f func()) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.last) >= t.duration {
		t.last = now
		f()
		return true
	}
	return false
}

// FormatDuration formats duration for human readability
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return d.String()
	}
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	return d.Round(time.Second).String()
}
