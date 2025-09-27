package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

// AtomicBool provides atomic boolean operations
type AtomicBool struct {
	value int32
}

func (b *AtomicBool) Set(value bool) {
	var i int32
	if value {
		i = 1
	}
	atomic.StoreInt32(&b.value, i)
}

func (b *AtomicBool) Get() bool {
	return atomic.LoadInt32(&b.value) != 0
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate       float64    // tokens per second
	capacity   float64    // maximum tokens
	tokens     float64    // current tokens
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed by the rate limiter
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.lastUpdate = now

	// Add tokens based on elapsed time
	rl.tokens += elapsed * rl.rate
	if rl.tokens > rl.capacity {
		rl.tokens = rl.capacity
	}

	// Check if we can take a token
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	for !rl.Allow() {
		time.Sleep(10 * time.Millisecond)
	}
}

// SizedWaitGroup extends sync.WaitGroup with size limiting
type SizedWaitGroup struct {
	size    int
	current chan struct{}
	wg      sync.WaitGroup
}

// NewSizedWaitGroup creates a wait group with maximum size
func NewSizedWaitGroup(size int) *SizedWaitGroup {
	return &SizedWaitGroup{
		size:    size,
		current: make(chan struct{}, size),
	}
}

// Add adds a goroutine, blocking if maximum size reached
func (swg *SizedWaitGroup) Add() {
	swg.current <- struct{}{}
	swg.wg.Add(1)
}

// Done signals a goroutine completion
func (swg *SizedWaitGroup) Done() {
	<-swg.current
	swg.wg.Done()
}

// Wait waits for all goroutines to complete
func (swg *SizedWaitGroup) Wait() {
	swg.wg.Wait()
}