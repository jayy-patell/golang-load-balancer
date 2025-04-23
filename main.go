package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"golang-load-balancer/algorithms"
	"golang-load-balancer/loadbalancer"
	"golang-load-balancer/servers"
)

func main() {
	// CLI flags
	algoFlag := flag.String("algo", "rr", "Load balancing strategy: rr, wrr, lc, ip")
	numFlag := flag.Int("n", 3, "Number of backend servers to spin up")
	weightsFlag := flag.String("weights", "", "Comma-separated weights for each server (used with wrr)")

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

	// Start dummy backend servers
	basePort := 8080
	log.Println("Starting backend servers...")
	go servers.RunServers(basePort, *numFlag)

	// Set up load balancer
	serverPool := loadbalancer.NewServerPool(strategyType)
	for i := 0; i < *numFlag; i++ {
		weight := i + 1 // Arbitrary weights for testing
		serverPool.AddBackend("http://localhost:", basePort+i, weight)
	}
	serverPool.InitStrategy(strategyType)

	// Start proxy
	loadbalancer.StartProxy(":8090", serverPool)
}

// go run main.go --algo=wrr --n=3 --weights=5,2,3
