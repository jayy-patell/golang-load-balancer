package algorithms

import (
	"sync"

	"golang-load-balancer/backend"
)

type WeightedRoundRobin struct {
	backends []*backend.Backend
	mutex    sync.Mutex
}

func NewWeightedRRPool(backends []*backend.Backend) *WeightedRoundRobin {
	return &WeightedRoundRobin{
		backends: backends,
	}
}

func (wrr *WeightedRoundRobin) GetStrategyType() StrategyType {
	return WeightedRoundRobinStrategy
}

func (wrr *WeightedRoundRobin) GetNextBackend() *backend.Backend {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	var totalWeight int
	var best *backend.Backend

	for _, b := range wrr.backends {
		if !b.IsAlive() {
			continue
		}

		b.CurrentWeight += b.Weight
		totalWeight += b.Weight

		if best == nil || b.CurrentWeight > best.CurrentWeight {
			best = b
		}
	}

	if best == nil {
		return nil
	}

	best.CurrentWeight -= totalWeight
	return best
}
