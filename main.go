package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"golang-load-balancer/algorithms"
	"golang-load-balancer/loadbalancer"
	"golang-load-balancer/servers"
)

func main() {
	// CLI flags
	algoFlag := flag.String("algo", "rr", "Load balancing strategy: rr, wrr, lc, ip")
	numFlag := flag.Int("n", 3, "Number of backend servers to spin up")
	weightsFlag := flag.String("weights", "", "Comma-separated weights for each server (used with wrr)")

	limiterFlag := flag.String("limiter", "none", "Rate limiter algorithm: none, token, fixed, leaky")
	rateFlag := flag.Int("rate", 5, "Allowed number of requests per second")
	burstFlag := flag.Int("burst", 5, "Burst size (only for token and leaky bucket)")

	flag.Parse()

	// Convert short algo names to StrategyType
	var strategyType algorithms.StrategyType
	switch *algoFlag {
	case "rr":
		strategyType = algorithms.RoundRobinStrategy
	case "wrr":
		strategyType = algorithms.WeightedRoundRobinStrategy
	case "lc":
		strategyType = algorithms.LeastConnectionsStrategy
	case "ip":
		strategyType = algorithms.IPHashStrategy
	default:
		log.Fatalf("Unknown strategy: %s. Use one of: rr, wrr, lc, ip", *algoFlag)
	}

	// Parse weights if provided
	var weights []int
	if *weightsFlag != "" {
		weightStrings := strings.Split(*weightsFlag, ",")
		for _, w := range weightStrings {
			parsed, err := strconv.Atoi(w)
			if err != nil {
				log.Fatalf("Invalid weight: %s", w)
			}
			weights = append(weights, parsed)
		}
	} else {
		// Default weight = 1 for each
		for i := 0; i < *numFlag; i++ {
			weights = append(weights, 1)
		}
	}

	basePort := 8080

	// Initialize server pool and backends
	serverPool := loadbalancer.NewServerPool(strategyType)
	for i := 0; i < *numFlag; i++ {
		serverPool.AddBackend("http://localhost:", basePort+i, weights[i])
	}
	serverPool.InitStrategy(strategyType)

	// 🔁 Start backend servers AFTER creating them
	log.Println("Starting backend servers...")
	go servers.RunServers(basePort, serverPool.GetBackends())

	// Start health checker
	go loadbalancer.StartHealthChecker(serverPool, 20*time.Second)

	// Start proxy server
	loadbalancer.StartProxy(":8090", serverPool, *limiterFlag, *rateFlag, *burstFlag)
}

// go run main.go -algo=rr -n=5
// go run main.go --algo=wrr --n=3 --weights=5,1,1
// go run main.go --algo=lc --n=3
// to test LC -> $for /L %i in (1,1,20) do start /B curl http://localhost:8090/loadbalancer
// go run main.go --algo=ip --n=3

// go run main.go -algo=rr -n=3 -limiter=token -rate=10 -burst=5
// go run main.go -algo=rr -n=3 -limiter=fixed -rate=5
// go run main.go -algo=rr -n=3 -limiter=leaky -rate=8 -burst=3
