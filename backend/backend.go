package backend

import (
	"log"
	"net/http"
	"net/url"
	"sync"
)

type Backend struct {
	URL               *url.URL
	Alive             bool
	Weight            int
	CurrentWeight     int
	ActiveConnections int

	mutex           sync.RWMutex // for Alive status
	ActiveConnMutex sync.RWMutex // for ActiveConnections
}

func (b *Backend) IsAlive() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Alive
}

func (b *Backend) IncrementConnections() {
	b.ActiveConnMutex.Lock()
	defer b.ActiveConnMutex.Unlock()
	b.ActiveConnections++
}

func (b *Backend) DecrementConnections() {
	b.ActiveConnMutex.Lock()
	defer b.ActiveConnMutex.Unlock()
	if b.ActiveConnections > 0 {
		b.ActiveConnections--
	}
}

func (b *Backend) GetConnections() int {
	b.ActiveConnMutex.RLock()
	defer b.ActiveConnMutex.RUnlock()
	return b.ActiveConnections
}

func CheckBackendHealth(u *url.URL) bool {
	resp, err := http.Get(u.String() + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed for %s", u.String())
		return false
	}
	return true
}
