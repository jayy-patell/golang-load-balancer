package main

import (
	"golang-load-balancer/servers"
)

func main() {
	servers.RunServers(5)
}