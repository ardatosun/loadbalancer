package serverpool

import (
	"loadbalancer/backends"
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

// ServerPool holds information about reachable backends
type ServerPool struct {
	backends []*backends.Backend
	current  uint32
}

// AddBackend adds a backend to the server pool
func (s *ServerPool) AddBackend(backend *backends.Backend) {
	s.backends = append(s.backends, backend)
}

// NextIndex atomically increases the counter and returns an index
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint32(&s.current, 1) % uint32(len(s.backends)))
}

// GetNextPeer returns the next active peer to take a connection
func (s *ServerPool) GetNextPeer() *backends.Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint32(&s.current, uint32(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

// MarkBackendStatus changes the status of a backend
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// HealthCheck pings the backends and updates their status
func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		status := "up"
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

// isBackendAlive checks whether a backend is alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error:", err)
		return false
	}
	_ = conn.Close()
	return true
}
