package serverpool

import (
	"context"
	"io"
	"loadbalancer/backends"
	"log"
	"net/http"
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

// GetNextPeer returns the next peer with the lowest latency
func (s *ServerPool) GetNextPeer() *backends.Backend {
	var selected *backends.Backend
	lowestLatency := time.Duration(1<<63 - 1) // Max int64

	for _, b := range s.backends {
		if b.IsAlive() {
			latency := b.GetLatency()
			if latency < lowestLatency {
				lowestLatency = latency
				selected = b
			}
		}
	}

	// Return the backend with the lowest latency
	if selected != nil {
		return selected
	}

	// Fallback to round-robin if no alive backend is found
	return s.backends[s.NextIndex()]
}

// GetBackends returns the list of backends in the pool
func (s *ServerPool) GetBackends() []*backends.Backend {
	return s.backends
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
	u.Path = "/health"
	timeout := 2 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Println("Unable to create the request context", err)
		panic(err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Site unreachable", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		log.Printf("Health check failed with status code: %d", res.StatusCode)
		return false
	}

	return true
}
