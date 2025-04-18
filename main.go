package main

import (
	"golang-load-balancer/loadbalancer"
	"golang-load-balancer/servers"
	"log"
	"time"
)

func main() {
	numServers := 5

	log.Println("Starting backend servers...")
	go servers.RunServers(numServers)

	log.Println("Waiting for servers to initialize...")
	time.Sleep(2 * time.Second)

	log.Println("Starting load balancer...")
	loadbalancer.MakeLoadBalancer(numServers)
}
