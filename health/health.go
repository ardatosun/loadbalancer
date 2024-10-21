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

// CheckHealth runs health checks for all backends at the given interval
func CheckHealth(pool *serverpool.ServerPool, interval time.Duration) {
	for {
		for _, b := range pool.GetBackends() {
			start := time.Now()
			alive := isBackendAlive(b.URL, interval)

			// Set backend status and latency
			b.SetAlive(alive)
			latency := time.Since(start)
			if alive {
				b.SetLatency(latency)
			} else {
				b.SetLatency(0)
			}

			status := "up"
			if !alive {
				status = "down"
			}
			log.Printf("Backend %s [%s] Latency: %v", b.URL, status, latency)
		}
		time.Sleep(interval)
	}
}

// isBackendAlive checks if a backend is alive by sending an HTTP request
func isBackendAlive(u *url.URL, timeout time.Duration) bool {
	u.Path = "/health"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return false
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		log.Printf("Health check failed for %s: %v", u.String(), err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	return true
}
