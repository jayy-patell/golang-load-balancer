package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"golang-load-balancer/algorithms"
	"golang-load-balancer/backend"
	"golang-load-balancer/ratelimiter"
)

// Custom ResponseWriter to detect when the response is sent
type responseWriter struct {
	http.ResponseWriter
	backend     *backend.Backend
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
		rw.backend.DecrementConnections()
		// log.Printf("Decremented. Server %s now has %d active connections", rw.backend.URL.String(), rw.backend.GetConnections())
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		// Ensure Decrement is called even if WriteHeader wasn't explicitly called
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// for per-client-ip rate limiting
var (
	limiterTypeGlobal string
	rateGlobal        int
	burstGlobal       int
	clientLimiters    = make(map[string]ratelimiter.Limiter)
	clientLimiterLock = sync.RWMutex{}
)

func allowRequest(r *http.Request) bool {
	clientIP := r.RemoteAddr
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		clientIP = ip
	} else if ip = r.Header.Get("X-Forwarded-For"); ip != "" {
		clientIP = strings.Split(ip, ",")[0]
	}

	clientLimiterLock.RLock()
	limiter, exists := clientLimiters[clientIP]
	clientLimiterLock.RUnlock()

	if !exists {
		clientLimiterLock.Lock()
		limiter = ratelimiter.NewLimiter(limiterTypeGlobal, rateGlobal, burstGlobal)
		clientLimiters[clientIP] = limiter
		clientLimiterLock.Unlock()
		log.Printf("Created new limiter for client %s", clientIP)
	}

	if !limiter.Allow(r) {
		log.Printf("Rate limit exceeded for client %s", clientIP)
		return false
	}
	return true
}

func StartProxy(port string, pool *ServerPool, limiterType string, rate int, burst int) {
	router := http.NewServeMux()

	limiterTypeGlobal = limiterType
	rateGlobal = rate
	burstGlobal = burst

	router.HandleFunc("/loadbalancer", func(w http.ResponseWriter, r *http.Request) {
		// check if client is requesting within limit
		if !allowRequest(r) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// serve next backend if allowed
		backend := pool.GetNextBackend()
		if backend == nil {
			log.Printf("No healthy backends available")
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		target := backend.URL
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.URL.Path = target.Path + r.URL.Path

		log.Printf("Forwarding request to: %s", target.String())

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "Service unavailable")
		}

		// Use custom writer only for LeastConnections
		if pool.GetStrategyType() == algorithms.LeastConnectionsStrategy {
			proxy.ServeHTTP(&responseWriter{ResponseWriter: w, backend: backend}, r)
		} else {
			proxy.ServeHTTP(w, r)
		}
	})

	log.Printf("Starting Load Balancer on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
