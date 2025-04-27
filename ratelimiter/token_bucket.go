package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity int // equal to burst
	tokens   int // available to use in bucket
	rate     int
	last     time.Time
	mutex    sync.Mutex
}

func NewTokenBucket(rate int, capacity int) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
		rate:     rate,
		last:     time.Now(),
	}
}

func (tb *TokenBucket) Allow(r *http.Request) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.last).Seconds()
	tb.last = now

	// Refill tokens based on elapsed time
	tb.tokens += int(elapsed * float64(tb.rate))
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}
