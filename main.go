package main

import (
	"context"
	"loadbalancer/backends"
	"loadbalancer/health"
	"loadbalancer/serverpool"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	Attempts int = iota
	Retry
)

var pool serverpool.ServerPool

func main() {
	// Read backend URLs from environment variable BACKEND_URLS
	backendURLs := strings.Split(strings.TrimSpace(os.Getenv("BACKEND_URLS")), ",")
	if len(backendURLs) == 0 {
		log.Fatal("BACKEND_URLS environment variable is not set")
	}

	// Parse backend URLs and add them to the server pool
	for _, backend := range backendURLs {
		url, err := url.Parse(backend)
		if err != nil {
			log.Fatal(err)
		}
		pool.AddBackend(&backends.Backend{URL: url, Alive: true})
	}

	// Start health checking in a separate goroutine
	go health.HealthCheck(&pool)

	// Handle incoming requests
	http.HandleFunc("/", handleRequest)
	log.Println("Load Balancer started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	loadBalance(w, r)
}

func loadBalance(w http.ResponseWriter, r *http.Request) {
	// Get the next backend server using round-robin
	backend := pool.GetNextPeer()
	if backend == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("[%s] %s\n", backend.URL.Host, e.Error())
		retries := GetRetryFromContext(r)
		if retries < 3 {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(r.Context(), Retry, retries+1)
				proxy.ServeHTTP(w, r.WithContext(ctx))
			}
			return
		}

		// After 3 retries, mark this backend as down
		pool.MarkBackendStatus(backend.URL, false)

		// If the same request routing for a few attempts with different backends, increase the count
		attempts := GetAttemptsFromContext(r)
		log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		ctx := context.WithValue(r.Context(), Attempts, attempts+1)
		loadBalance(w, r.WithContext(ctx)) // Retry using lb function
	}

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

// GetAttemptsFromContext returns the number of attempts for a request
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 0
}

// GetRetryFromContext returns the number of retries for a request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}
