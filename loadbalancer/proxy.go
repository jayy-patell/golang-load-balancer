package loadbalancer

import (
	"fmt"
	"golang-load-balancer/algorithms"
	"log"
	"net/http"
	"net/http/httputil"
)

func StartProxy(port string, pool *ServerPool) {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backend := pool.GetNextBackend()
			if backend == nil {
				log.Printf("No healthy backends available")
				req.URL = nil
				return
			}

			// Only decrement for Least Connections strategy
			if pool.GetStrategyType() == algorithms.LeastConnectionsStrategy {
				defer log.Printf("Decremented.. Active Connections: %d", backend.ActiveConnections)
				defer backend.DecrementConnections() // Decrement after request is handled
				defer log.Printf("Decremented.. Active Connections: %d", backend.ActiveConnections)
			}

			target := backend.URL
			log.Printf("Forwarding request to: %s", target.String())

			req.URL.Scheme = target.Scheme
			log.Printf("Req URL Scheme: %s, Target URL Scheme: %s", req.URL.Scheme, target.Scheme)
			req.URL.Host = target.Host
			log.Printf("Req URL Host: %s, Target URL Host: %s", req.URL.Host, target.Host)
			req.URL.Path = target.Path + req.URL.Path
			log.Printf("Req URL Path: %s, Target URL Path: %s", req.URL.Path, target.Path)

			// or http.Redirect
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "Service unavailable")
		},
	}

	router := http.NewServeMux()
	router.Handle("/loadbalancer", proxy)

	log.Printf("Starting Load Balancer on %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
