package servers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang-load-balancer/backend"
)

// RunServers starts multiple dummy servers using the given backend instances
func RunServers(basePort int, backends []*backend.Backend) {
	if len(backends) > 10 {
		log.Fatal("Amount of servers cannot exceed 10")
	}

	// Waitgroup to track when all servers are started
	var wg sync.WaitGroup
	wg.Add(len(backends))

	for i, b := range backends {
		go startServer(basePort+i, i, b, &wg)
	}

	// Wait for all servers to start
	wg.Wait()
	log.Printf("All %d servers started successfully", len(backends))
}

// startServer sets up and runs a dummy server on a given port
func startServer(port int, serverID int, b *backend.Backend, wg *sync.WaitGroup) {
	defer wg.Done()

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
		w.WriteHeader(http.StatusOK)
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
