package servers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// RunServers starts multiple dummy servers from a base port
func RunServers(basePort int, amount int) {
	if amount > 10 {
		log.Fatal("Amount of servers cannot exceed 10")
	}

	// Waitgroup to track when all servers are started
	var wg sync.WaitGroup
	wg.Add(amount)

	// Start each server with its own index/port
	for i := 0; i < amount; i++ {
		go startServer(basePort+i, i, &wg)
	}

	// Wait for all servers to start
	wg.Wait()
	log.Printf("All %d servers started successfully", amount)
}

// startServer sets up and runs a dummy server on a given port
func startServer(port int, serverID int, wg *sync.WaitGroup) {
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
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server %d is healthy", serverID)
	})

	router.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Server %d is shutting down...", serverID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server %d shutting down", serverID)

		go func() {
			time.Sleep(1 * time.Second)
			server.Close()
		}()
	})

	// Start server (this blocks until the server stops)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server %d error: %v", serverID, err)
	}
}
