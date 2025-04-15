package servers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type ServerList struct {
	Ports []int
}

func (s *ServerList) populate(amount int) {
	if amount > 8 {
		log.Fatal("Amount of ports cannot exceed 8")
	}
	for x:=0; x<amount; x++ {
		s.Ports = append(s.Ports, x)
	}
}

func (s* ServerList) pop() int {
	port := s.Ports[0]
	s.Ports = s.Ports[1:]
	return port
}

func RunServers(amount int) {
	// ServerList Object
	var myServerList ServerList
	myServerList.populate(amount)

	// Waitgroup 
	var wg sync.WaitGroup
	wg.Add(amount)
	defer wg.Wait()

	for range amount {
		go makeServers(&myServerList, &wg)
	}
}

func makeServers(sl *ServerList, wg *sync.WaitGroup) {
	//Router 
	r := http.NewServeMux()

	defer wg.Done()

	// Port
	port := sl.pop()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Server %d", port)
	})
	server := http.Server{
		Addr: fmt.Sprintf(":808%d",port),
		Handler: r,
	}
	server.ListenAndServe()
}

//for concurrency- there are channels and waitgroups