package serverpool

import (
	"loadbalancer/backends"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type ServerPool struct {
	backends              []*backends.Backend
	current               uint32 // To maintain round-robin state for backends
	GlobalRateLimitPerSec int    // New global rate limit
}

// GetBackends returns all backends.
func (s *ServerPool) GetBackends() []*backends.Backend {
	return s.backends
}

// AddBackend adds a backend to the server pool and sets up the rate limiter
func (s *ServerPool) AddBackend(backend *backends.Backend, rateLimitPerSec int) {
	backend.RequestLimiter = rate.NewLimiter(rate.Limit(rateLimitPerSec), rateLimitPerSec)
	s.backends = append(s.backends, backend)
}

// GetNextPeer returns the next available backend based on latency or round-robin
func (s *ServerPool) GetNextPeer() *backends.Backend {
	var selected *backends.Backend
	lowestLatency := time.Duration(1<<63 - 1) // Max int64 to find the backend with the lowest latency

	for _, b := range s.backends {
		if b.IsAlive() {
			latency := b.GetLatency()
			if latency < lowestLatency {
				lowestLatency = latency
				selected = b
			}
		}
	}

	if selected != nil {
		log.Printf("Selected backend %s with lowest latency: %v", selected.URL, lowestLatency)
		return selected
	}

	index := s.NextIndex()
	log.Printf("Falling back to round-robin. Selected backend: %s", s.backends[index].URL)
	return s.backends[index]
}

// CheckRateLimit checks if a backend can handle more requests, otherwise returns HTTP 429
func (s *ServerPool) CheckRateLimit(backend *backends.Backend, w http.ResponseWriter) bool {
	if !backend.AllowRequest() {
		http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
		return false
	}
	return true
}

// NextIndex atomically increases the counter and returns an index for round-robin load balancing
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint32(&s.current, 1) % uint32(len(s.backends)))
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
	timeout := 2 * time.Second

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
