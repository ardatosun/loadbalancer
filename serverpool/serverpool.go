package serverpool

import (
	"loadbalancer/backends"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

// ServerPool holds information about the list of backend servers and their current state
type ServerPool struct {
	backends []*backends.Backend
	current  uint32 // To maintain round-robin state for backends
}

// AddBackend adds a backend to the server pool
func (s *ServerPool) AddBackend(backend *backends.Backend) {
	s.backends = append(s.backends, backend)
}

// NextIndex atomically increases the counter and returns an index for round-robin load balancing
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint32(&s.current, 1) % uint32(len(s.backends)))
}

// GetNextPeer returns the next available backend based on latency. If no backend is alive, it falls back to round-robin.
func (s *ServerPool) GetNextPeer() *backends.Backend {
	var selected *backends.Backend
	lowestLatency := time.Duration(1<<63 - 1) // Max int64 to find the backend with the lowest latency

	for _, b := range s.backends {
		// Only consider backends that are alive
		if b.IsAlive() {
			latency := b.GetLatency()
			if latency < lowestLatency {
				lowestLatency = latency
				selected = b
			}
		}
	}

	// If no alive backends found, fall back to round-robin
	if selected != nil {
		log.Printf("Selected backend %s with lowest latency: %v", selected.URL, lowestLatency)
		return selected
	}

	// Fallback to round-robin selection
	index := s.NextIndex()
	log.Printf("Falling back to round-robin. Selected backend: %s", s.backends[index].URL)
	return s.backends[index]
}

// GetBackends returns the list of backends in the server pool
func (s *ServerPool) GetBackends() []*backends.Backend {
	return s.backends
}

// MarkBackendStatus marks a backend's status as alive or down based on health checks
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			status := "up"
			if !alive {
				status = "down"
			}
			log.Printf("Backend %s marked as %s", b.URL, status)
			break
		}
	}
}

// HealthCheck runs the health checks for all backends in the pool and updates their statuses
func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		status := "up"
		if !alive {
			status = "down"
		}
		log.Printf("Backend %s is %s", b.URL, status)
	}
}

// isBackendAlive checks whether a backend is reachable and responds with an HTTP 200 status code
func isBackendAlive(u *url.URL) bool {
	u.Path = "/health"
	timeout := 2 * time.Second // Health check timeout

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("Error creating health check request: %v", err)
		return false
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err != nil {
			log.Printf("Error during health check for %s: %v", u, err)
		} else {
			log.Printf("Health check failed for %s with status code: %d", u, resp.StatusCode)
		}
		return false
	}

	defer resp.Body.Close()
	return true
}
