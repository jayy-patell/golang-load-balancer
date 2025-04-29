package algorithms

import (
	"hash/crc32"
	"log"
	"net"
	"net/http"
	"strings"

	"golang-load-balancer/backend"
)

type IPHash struct {
	backends []*backend.Backend
}

func NewIPHash(backends []*backend.Backend) *IPHash {
	return &IPHash{backends: backends}
}

func (ih *IPHash) GetStrategyType() StrategyType {
	return IPHashStrategy
}

func (h *IPHash) UpdateBackends(backends []*backend.Backend) {
	h.backends = backends
}

// GetNextBackend returns the same backend for same hash value
func (ip *IPHash) GetNextBackend() *backend.Backend {
	// Get real client IP from request context (you would get this in a real scenario)
	clientIP := "192.168.1.100"

	// Hash the client IP address using CRC32 to determine the backend
	hash := crc32.ChecksumIEEE([]byte(clientIP))
	index := int(hash % uint32(len(ip.backends)))

	// Return the selected backend
	selectedBackend := ip.backends[index]
	log.Printf("IP Hashing selected backend: %s for client IP: %s", selectedBackend.URL.String(), clientIP)
	return selectedBackend
}

// for sticky session for each client
// func (ih *IPHash) GetNextBackend(r *http.Request) *backend.Backend {
// 	// Get client IP from request
// 	clientIP := getClientIP(r)

// 	// Use a simple hash function (you can replace this with a more complex one if needed)
// 	hash := hashIP(clientIP)

// 	// Get backend based on hashed value
// 	index := hash % len(ih.backends)
// 	backend := ih.backends[index]

// 	log.Printf("Routing request from IP %s to backend %s", clientIP, backend.URL.String())
// 	return backend
// }

// getClientIP retrieves the client IP from the HTTP request.
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header if behind a proxy
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// The X-Forwarded-For header can have multiple IPs, use the first one
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to RemoteAddr if no X-Forwarded-For header
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("Error parsing RemoteAddr:", err)
		return ""
	}

	return host
}

// A simple hash function (you can replace it with a more complex hash like MD5/SHA256 if needed)
func hashIP(clientIP string) int {
	var hash int
	for i := 0; i < len(clientIP); i++ {
		hash = 31*hash + int(clientIP[i])
	}
	return hash
}

/*
In the context of load balancing and IP Hashing, clientIP refers to the IP address of the client
(i.e., the end-user or requesting machine) that is making the HTTP request to the load balancer.
This IP is used to ensure that requests from the same client are always routed to the same backend server
(a concept known as "sticky sessions" or "session persistence").
*/
