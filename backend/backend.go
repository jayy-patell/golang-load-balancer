package backend

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL               *url.URL
	Alive             bool
	Weight            int
	CurrentWeight     int
	ActiveConnections int

	mutex           sync.RWMutex // for Alive status
	ActiveConnMutex sync.RWMutex // for ActiveConnections
}

func (b *Backend) IsAlive() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Alive
}

func (b *Backend) SetAlive(alive bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Alive = alive
}

func (b *Backend) IncrementConnections() {
	b.ActiveConnMutex.Lock()
	defer b.ActiveConnMutex.Unlock()
	b.ActiveConnections++
}

func (b *Backend) DecrementConnections() {
	b.ActiveConnMutex.Lock()
	defer b.ActiveConnMutex.Unlock()
	if b.ActiveConnections > 0 {
		b.ActiveConnections--
	}
}

func (b *Backend) GetConnections() int {
	b.ActiveConnMutex.RLock()
	defer b.ActiveConnMutex.RUnlock()
	return b.ActiveConnections
}

func CheckBackendHealth(u *url.URL) bool {
	resp, err := http.Get(u.String() + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed for %s", u.String())
		return false
	}
	return true
}

// RunServers starts multiple dummy servers using the given backend instances
func RunServers(basePort int, backends []*Backend) {
	if len(backends) > 10 {
		log.Fatal("Amount of servers cannot exceed 10")
	}

	// Waitgroup to track when all servers are started
	var wg sync.WaitGroup
	wg.Add(len(backends))

	for i, b := range backends {
		go StartServer(basePort+i, i, b, &wg)
	}

	// Wait for all servers to start
	wg.Wait()
	log.Printf("All %d servers started successfully", len(backends))
}

// startServer sets up and runs a dummy server on a given port
func StartServer(port int, serverID int, b *Backend, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	// Create router
	router := http.NewServeMux()

	// Configure route handler
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Server %d handling request %s", serverID, r.URL.Path)

		// Simulate heavy load
		time.Sleep(2 * time.Second)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Response from Server %d", serverID)
	})

	// Configure server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server %d on %s", serverID, addr)

	server := http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Handler functions
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if b.IsAlive() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Server %d is healthy", serverID)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Server %d is unhealthy", serverID)
		}
	})

	router.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Server %d is shutting down...", serverID)
		fmt.Fprintf(w, "Server %d shutting down", serverID)

		b.SetAlive(false) // mark as unhealthy

		go func() {
			time.Sleep(1 * time.Second) // wait for inflight requests
			server.Close()
		}()
	})

	// Start server (this blocks until the server stops)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server %d error: %v", serverID, err)
	}
}
