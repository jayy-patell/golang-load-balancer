package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
)

var (
	baseURL = "http://localhost:808"
)

// ServerPool holds information about available backend servers
type ServerPool struct {
	backends []*Backend
	current  int
	mutex    sync.Mutex
}

// Backend represents a single backend server
type Backend struct {
	URL   *url.URL
	Alive bool
	mutex sync.RWMutex
}

func (b *Backend) IsAlive() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Alive
}

// GetNextBackend returns the next available backend server
func (s *ServerPool) GetNextBackend() *Backend {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	n := len(s.backends)
	for i := 0; i < n; i++ {
		index := (s.current + i) % n
		backend := s.backends[index]

		if checkBackendHealth(backend.URL) {
			// Set the next backend as the starting point for round-robin selection
			s.current = (index + 1) % n

			// Return the current healthy backend to handle the request
			return backend
		}
	}
	return nil // No healthy server found
}

// LoadBalancerHandler creates a custom director for the reverse proxy
func LoadBalancerHandler(pool *ServerPool) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		backend := pool.GetNextBackend()
		if backend == nil {
			log.Printf("No healthy backends available")
			req.URL = nil
			return
		}

		target := backend.URL
		log.Printf("Forwarding request to: %s", target.String())

		req.URL.Scheme = target.Scheme
		log.Printf("Req URL Scheme: %s, Target URL Scheme: %s", req.URL.Scheme, target.Scheme)
		req.URL.Host = target.Host
		log.Printf("Req URL Host: %s, Target URL Host: %s", req.URL.Host, target.Host)
		req.URL.Path = target.Path + req.URL.Path
		log.Printf("Req URL Path: %s, Target URL Path: %s", req.URL.Path, target.Path)
	}

	return &httputil.ReverseProxy{
		Director: director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Error forwarding request: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Service Unavailable: %v", err)
		},
	}
}

// checkBackendHealth does a single /health call to validate a backend
func checkBackendHealth(url *url.URL) bool {
	resp, err := http.Get(url.String() + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed for %s", url)
		return false
	}
	return true
}

func MakeLoadBalancer(amount int) {
	// Initialising the pool
	serverPool := &ServerPool{
		backends: make([]*Backend, 0),
		current:  0,
	}

	for i := 0; i < amount; i++ {
		serverPool.backends = append(serverPool.backends, addEnpoint(baseURL, i))
	}

	// Create a single reverse proxy
	proxy := LoadBalancerHandler(serverPool)

	// Server + Router
	router := http.NewServeMux()
	server := http.Server{
		Addr:    ":8090",
		Handler: router,
	}

	// Route all traffic through our load balancer
	router.Handle("/loadbalancer", proxy)

	// Listen and Serve
	log.Printf("Load balancer starting on :8090")
	log.Fatal(server.ListenAndServe())
}

func addEnpoint(endpoint string, idx int) *Backend {
	link := endpoint + strconv.Itoa(idx)
	url, err := url.Parse(link)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", link, err)
		return nil
	}
	log.Printf("Added endpoint: %s", url.String())

	resp, err := http.Get(url.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return &Backend{URL: url, Alive: false}
	}
	return &Backend{URL: url, Alive: true}
}
