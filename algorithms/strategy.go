package algorithms

import (
	"log"

	"golang-load-balancer/backend"
)

type StrategyType string

const (
	RoundRobinStrategy         StrategyType = "round_robin"
	WeightedRoundRobinStrategy StrategyType = "weighted_round_robin"
	LeastConnectionsStrategy   StrategyType = "least_connections"
	IPHashStrategy             StrategyType = "ip_hash"
)

type Strategy interface {
	GetNextBackend() *backend.Backend
	GetStrategyType() StrategyType
}

// Any struct that has a GetNextBackend() method with this exact signature can be treated as a Strategy.

func NewStrategy(strategy StrategyType, backends []*backend.Backend) Strategy {
	switch strategy {
	case RoundRobinStrategy:
		return NewRoundRobin(backends)
	case WeightedRoundRobinStrategy:
		return NewWeightedRRPool(backends)
	case LeastConnectionsStrategy:
		return NewLeastConnections(backends)
	case IPHashStrategy:
		return NewIPHash(backends)

	default:
		log.Fatal("Invalid algorithm. Use: rr, wrr, ip, lc")
		return nil
	}
}
