package loadbalancer

import (
	"log"
	"net/url"
	"strconv"
	"sync"

	"golang-load-balancer/algorithms"
	"golang-load-balancer/backend"
)

type ServerPool struct {
	backends []*backend.Backend
	strategy algorithms.Strategy
	mutex    sync.Mutex
}

func NewServerPool(strategyType algorithms.StrategyType) *ServerPool {
	return &ServerPool{
		backends: []*backend.Backend{},
		strategy: nil,
	}
}

func (s *ServerPool) InitStrategy(strategyType algorithms.StrategyType) {
	s.strategy = algorithms.NewStrategy(strategyType, s.backends)
}

func (s *ServerPool) GetBackends() []*backend.Backend {
	return s.backends
}

func (s *ServerPool) GetNextBackend() *backend.Backend {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.strategy == nil {
		return nil
	}
	return s.strategy.GetNextBackend()
}

func (s *ServerPool) GetStrategyType() algorithms.StrategyType {
	// Returns the current strategy type
	if s.strategy != nil {
		return s.strategy.GetStrategyType()
	}
	return ""
}

func (s *ServerPool) AddBackend(endpoint string, idx int, weight int) {
	link := endpoint + strconv.Itoa(idx)
	parsedURL, err := url.Parse(link)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", link, err)
		return
	}

	b := &backend.Backend{
		URL:               parsedURL,
		Alive:             true, // assume
		Weight:            weight,
		CurrentWeight:     0,
		ActiveConnections: 0,
	}
	s.backends = append(s.backends, b)
	log.Printf("Added backend: %s", b.URL.String())
}
