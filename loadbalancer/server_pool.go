package loadbalancer

import (
	"fmt"
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

func (s *ServerPool) AddBackendUsingIndex(endpoint string, idx int, weight int) {
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

// Dynamic AddBackend at runtime
func (s *ServerPool) AddBackendDynamic(backendURL string, weight int) (*backend.Backend, error) {
	parsedURL, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	b := &backend.Backend{
		URL:               parsedURL,
		Alive:             true,
		Weight:            weight,
		CurrentWeight:     0,
		ActiveConnections: 0,
	}
	s.backends = append(s.backends, b)

	// Important: Reinitialize strategy because number of backends changed
	if s.strategy != nil {
		s.strategy.UpdateBackends(s.backends)
	}

	log.Printf("Dynamically added backend: %s", backendURL)
	return b, nil
}

// Dynamic RemoveBackend at runtime
func (s *ServerPool) RemoveBackendDynamic(backendURL string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, b := range s.backends {
		if b.URL.String() == backendURL {
			s.backends = append(s.backends[:i], s.backends[i+1:]...)

			// Important: Reinitialize strategy because number of backends changed
			if s.strategy != nil {
				s.strategy.UpdateBackends(s.backends)
			}

			log.Printf("Dynamically removed backend: %s", backendURL)
			return nil
		}
	}
	return fmt.Errorf("backend %s not found", backendURL)
}
