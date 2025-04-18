package servers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type ServerList struct {
	Ports   []int
	current int
	mutex   sync.Mutex
}

func (s *ServerList) populate(amount int) {
	if amount > 8 {
		log.Fatal("Amount of ports cannot exceed 8")
	}
	for x := 0; x < amount; x++ {
		s.Ports = append(s.Ports, x)
	}
}

// getNext returns the next port
func (s *ServerList) getNext() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	port := s.Ports[s.current]
	s.current = (s.current + 1) % len(s.Ports)
	return port
}

func RunServers(amount int) {
	// ServerList Object
	var myServerList ServerList
	myServerList.populate(amount)

	// Waitgroup
	var wg sync.WaitGroup
	wg.Add(amount)

	for i := 0; i < amount; i++ {
		go makeServer(&myServerList, &wg)
	}

	// Wait for all servers to start
	wg.Wait()
	log.Printf("All %d servers are running", amount)
}

func makeServer(sl *ServerList, wg *sync.WaitGroup) {
	defer wg.Done()

	// Router
	router := http.NewServeMux()

	// Get port from server list
	port := sl.getNext()

	// Configure route handler
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Server %d handling request %s", port, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Response from Server %d", port)
	})

	// Configure server
	server := http.Server{
		Addr:    fmt.Sprintf(":808%d", port),
		Handler: router,
	}
	log.Printf("Started server on %s", server.Addr)

	// Start server (this blocks until the server stops)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server %d error: %v", port, err)
	}
}
