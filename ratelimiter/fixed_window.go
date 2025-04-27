package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

type FixedWindow struct {
	rate      int // max #requests allowed in a window 
	count     int // tracks #requests in one time window
	startTime time.Time
	mutex     sync.Mutex
}

func NewFixedWindow(rate int) *FixedWindow {
	return &FixedWindow{
		rate:      rate,
		startTime: time.Now(),
	}
}

func (fw *FixedWindow) Allow(r *http.Request) bool {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	now := time.Now()
	if now.Sub(fw.startTime) > time.Second {
		fw.startTime = now
		fw.count = 0
	}

	if fw.count < fw.rate {
		fw.count++
		return true
	}
	return false
}
