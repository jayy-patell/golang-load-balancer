package loadbalancer

import (
	"time"

	"golang-load-balancer/backend"
)

func StartHealthChecker(pool *ServerPool, interval time.Duration) {
	go func() {
		for {
			for _, b := range pool.GetBackends() {
				alive := backend.CheckBackendHealth(b.URL)
				b.SetAlive(alive)
			}
			time.Sleep(interval)
		}
	}()
}
