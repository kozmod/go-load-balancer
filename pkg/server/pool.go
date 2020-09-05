package server

import (
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

// Pool holds information about reachable servers
type Pool struct {
	servers []*Server
	current uint64
}

// Add a Server to the server pool
func (s *Pool) Add(backend *Server) {
	s.servers = append(s.servers, backend)
}

// MarkBackendStatus changes a status of a backend
func (s *Pool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.servers {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// GetNextPeer returns next active peer to take a connection
func (s *Pool) GetNextPeer() *Server {
	// loop entire servers to find out an Alive backend
	var (
		next = s.nextIndex()
		sln  = len(s.servers)
		l    = sln + next
	)
	for i := next; i < l; i++ {
		idx := i % sln
		if s.servers[idx].IsAlive() { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.servers[idx]
		}
	}
	return nil
}

// nextIndex atomically increase the counter and return an index
func (s *Pool) nextIndex() int {
	sln := len(s.servers)
	idx := int(atomic.AddUint64(&s.current, uint64(1)) % uint64(sln))
	if idx > sln-1 {
		idx = 0
	}
	return idx
}

// HealthCheck pings the servers and update the status
func (s *Pool) HealthCheck() {
	for _, b := range s.servers {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

// isBackendAlive checks whether a backend is Alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("Close error: ", err)
		}
	}()
	return true
}
