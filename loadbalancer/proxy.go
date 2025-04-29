package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
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
	if limiterTypeGlobal == "none" {
		return true // no limiting needed
	}

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

// AddBackend adds a backend to the pool and starts a dummy server
func addBackend(w http.ResponseWriter, r *http.Request, pool *ServerPool) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse parameters
	rawURL := r.URL.Query().Get("url")
	weightStr := r.URL.Query().Get("weight")
	if rawURL == "" {
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	weight := 1
	if weightStr != "" {
		w, err := strconv.Atoi(weightStr)
		if err == nil && w > 0 {
			weight = w
		}
	}

	backendURL, err := url.Parse(rawURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	portStr := backendURL.Port()
	if portStr == "" {
		http.Error(w, "Port must be specified", http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid port number", http.StatusBadRequest)
		return
	}

	// Add backend to pool
	newBackend, err := pool.AddBackendDynamic(rawURL, weight)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add backend: %v", err), http.StatusInternalServerError)
		return
	}

	// Spin up server
	go backend.StartServer(port, port%10, newBackend, nil)

	// Respond success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Backend %s added with weight %d", rawURL, weight)
}

// RemoveBackend removes a backend from the pool by its URL
func removeBackend(w http.ResponseWriter, r *http.Request, pool *ServerPool) {
	// Parse parameters (URL)
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	// Call the /shutdown endpoint using POST
	shutdownURL := fmt.Sprintf("%s/shutdown", url)
	resp, err := http.Post(shutdownURL, "application/json", nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Failed to call shutdown on backend: %v", err), http.StatusBadGateway)
		return
	}

	// Remove the backend from the pool
	err = pool.RemoveBackendDynamic(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove backend: %v", err), http.StatusBadRequest)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Backend removed: %s", url)
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

	router.HandleFunc("/admin/addBackend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}
		addBackend(w, r, pool)
	})

	router.HandleFunc("/admin/removeBackend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}
		removeBackend(w, r, pool)
	})

	log.Printf("Starting Load Balancer on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
