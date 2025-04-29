package algorithms

import (
	"sync"

	"golang-load-balancer/backend"
)

type RoundRobin struct {
	backends []*backend.Backend
	current  int
	mutex    sync.Mutex
}

func NewRoundRobin(backends []*backend.Backend) *RoundRobin {
	return &RoundRobin{
		backends: backends, current: 0,
	}
}

func (rr *RoundRobin) GetStrategyType() StrategyType {
	return RoundRobinStrategy
}

func (r *RoundRobin) UpdateBackends(backends []*backend.Backend) {
	r.backends = backends
}

func (rr *RoundRobin) GetNextBackend() *backend.Backend {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	n := len(rr.backends)
	for i := 0; i < n; i++ {
		index := (rr.current + i) % n
		backend := rr.backends[index]

		if backend.IsAlive() {
			// Set the next backend as the starting point for round-robin selection
			rr.current = (index + 1) % n

			// Return the current healthy backend to handle the request
			return backend
		}
	}
	return nil // No healthy server found
}
