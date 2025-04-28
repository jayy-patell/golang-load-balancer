package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

type LeakyBucket struct {
	capacity  int // max #requests bucket can hold
	rate      float64 // allowed rate at which requests can be processed
	water     float64 // current #requests in bucket
	lastCheck time.Time
	mutex     sync.Mutex
}

func NewLeakyBucket(rate int, capacity int) *LeakyBucket {
	return &LeakyBucket{
		capacity:  capacity,
		rate:      float64(rate),
		lastCheck: time.Now(),
	}
}

func (lb *LeakyBucket) Allow(r *http.Request) bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastCheck).Seconds()
	lb.lastCheck = now

	lb.water -= elapsed * lb.rate
	if lb.water < 0 {
		lb.water = 0
	}

	if lb.water < float64(lb.capacity) {
		lb.water++
		return true
	}
	return false
}
