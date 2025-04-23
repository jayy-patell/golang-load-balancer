package algorithms

import (
	"log"

	"golang-load-balancer/backend"
)

type LeastConnections struct {
	backends []*backend.Backend
}

func NewLeastConnections(backends []*backend.Backend) *LeastConnections {
	return &LeastConnections{backends: backends}
}

func (lc *LeastConnections) GetStrategyType() StrategyType {
	return LeastConnectionsStrategy
}

func (lc *LeastConnections) GetNextBackend() *backend.Backend {
	var best *backend.Backend
	minConnections := -1

	for _, b := range lc.backends {
		if !backend.CheckBackendHealth(b.URL) {
			continue
		}
		b.ActiveConnMutex.RLock()
		curConnections := b.ActiveConnections
		b.ActiveConnMutex.RUnlock()
		if minConnections == -1 || curConnections < minConnections {
			best = b
			minConnections = curConnections
		}
	}
	if best != nil {
		best.IncrementConnections()
		log.Printf("Server %s has %d active connections", best.URL.String(), best.ActiveConnections)
	}
	return best
}
