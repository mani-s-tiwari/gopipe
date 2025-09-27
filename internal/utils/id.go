package utils

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"time"
)

var (
	// Atomic counter for unique ID generation
	idCounter uint64
)

// GenerateID generates a unique ID with timestamp and random component
func GenerateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	counter := atomic.AddUint64(&idCounter, 1)
	randomPart := generateRandomString(8)
	
	return fmt.Sprintf("%s_%d_%d_%s", prefix, timestamp, counter, randomPart)
}

// GenerateTaskID generates a task-specific ID
func GenerateTaskID() string {
	return GenerateID("task")
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b) // Ignoring error for simplicity
	
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// ShortID generates a shorter ID for logging
func ShortID() string {
	return generateRandomString(6)
}