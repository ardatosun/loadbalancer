package health

import (
	"context"
	"io"
	"loadbalancer/serverpool"
	"log"
	"net/http"
	"net/url"
	"time"
)

// HealthCheck runs the health check for all backends in the server pool and measures latency
func HealthCheck(pool *serverpool.ServerPool) {
	for _, b := range pool.GetBackends() {
		start := time.Now() // Start measuring time
		alive := isBackendAlive(b.URL)
		latency := time.Since(start) // Calculate latency

		// Set the backend's status and latency
		b.SetAlive(alive)
		if alive {
			b.SetLatency(latency) // Update latency if the backend is alive
		} else {
			b.SetLatency(0) // Reset latency if backend is down
		}

		status := "up"
		if !alive {
			status = "down"
		}
		log.Printf("Backend: %s [%s] - Latency: %s\n", b.URL.String(), status, latency)
	}
}

// isBackendAlive checks whether a backend is alive and reachable
func isBackendAlive(u *url.URL) bool {
	u.Path = "/health"
	timeout := 2 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Println("Unable to create the request context", err)
		return false
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

	// Backend is considered alive if status code is 200
	if res.StatusCode != http.StatusOK {
		log.Printf("Health check failed for %s with status code: %d", u.String(), res.StatusCode)
		return false
	}

	return true
}
