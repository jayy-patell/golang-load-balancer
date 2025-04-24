package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"golang-load-balancer/algorithms"
	"golang-load-balancer/backend"
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

func StartProxy(port string, pool *ServerPool) {
	router := http.NewServeMux()

	router.HandleFunc("/loadbalancer", func(w http.ResponseWriter, r *http.Request) {
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
