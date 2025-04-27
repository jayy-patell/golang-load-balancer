package ratelimiter

import "net/http"

type Limiter interface {
	Allow(r *http.Request) bool
}

func NewLimiter(algo string, rate int, burst int) Limiter {
	switch algo {
	case "token":
		return NewTokenBucket(rate, burst)
	case "fixed":
		return NewFixedWindow(rate)
	case "leaky":
		return NewLeakyBucket(rate, burst)
	default:
		return nil
	}
}
