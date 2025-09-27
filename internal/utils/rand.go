package utils

import (
	"math/rand"
	"sync"
	"time"
)

var (
	globalRand *rand.Rand
	randMu     sync.Mutex
	randOnce   sync.Once
)

// initRand initializes the global random generator
func initRand() {
	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// SecureRandomString generates a cryptographically secure random string
func SecureRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

// RandomInt returns a random integer in [min, max]
func RandomInt(min, max int) int {
	randOnce.Do(initRand)
	randMu.Lock()
	defer randMu.Unlock()
	
	return min + globalRand.Intn(max-min+1)
}

// RandomDuration returns a random duration in [min, max]
func RandomDuration(min, max time.Duration) time.Duration {
	randOnce.Do(initRand)
	randMu.Lock()
	defer randMu.Unlock()
	
	nanos := min.Nanoseconds() + globalRand.Int63n(max.Nanoseconds()-min.Nanoseconds()+1)
	return time.Duration(nanos) * time.Nanosecond
}

// WeightedRandom selects a random item based on weights
func WeightedRandom[T any](items []T, weights []float64) T {
	if len(items) != len(weights) {
		panic("items and weights must have same length")
	}
	
	randOnce.Do(initRand)
	randMu.Lock()
	defer randMu.Unlock()
	
	// Calculate cumulative weights
	cumulative := make([]float64, len(weights))
	sum := 0.0
	for i, w := range weights {
		sum += w
		cumulative[i] = sum
	}
	
	// Select random value
	r := globalRand.Float64() * sum
	for i, c := range cumulative {
		if r <= c {
			return items[i]
		}
	}
	
	return items[len(items)-1]
}

// Shuffle randomly shuffles a slice in place
func Shuffle[T any](slice []T) {
	randOnce.Do(initRand)
	randMu.Lock()
	defer randMu.Unlock()
	
	globalRand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}