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
	URL *url.URL
}

// GetNextBackend returns the next available backend server
func (s *ServerPool) GetNextBackend() *Backend {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backend := s.backends[s.current]
	s.current = (s.current + 1) % len(s.backends) //update current with next server
	return backend                                //return the previous server
}

// LoadBalancerHandler creates a custom director for the reverse proxy
func LoadBalancerHandler(pool *ServerPool) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		backend := pool.GetNextBackend()
		target := backend.URL

		log.Printf("Forwarding request to: %s", target.String())

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path + req.URL.Path

		// if target.Host != "" {
		// 	req.Host = target.Host
		// }
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

func MakeLoadBalancer(amount int) {
	serverPool := &ServerPool{
		backends: make([]*Backend, 0),
		current:  0,
	}

	for i := 0; i < amount; i++ {
		serverPool.backends = append(serverPool.backends, createEndpoint(baseURL, i))
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

func createEndpoint(endpoint string, idx int) *Backend {
	link := endpoint + strconv.Itoa(idx)
	url, err := url.Parse(link)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", link, err)
		return nil
	}
	log.Printf("Added endpoint: %s", url.String())
	return &Backend{URL: url}
}
