package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"loadbalancer/backends"
	"loadbalancer/health"
	"loadbalancer/serverpool"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}

const (
	Attempts int = iota
	Retry
)

var pool serverpool.ServerPool

func main() {
	// Read backend URLs from environment variable BACKEND_URLS
	backendURLs := strings.Split(strings.TrimSpace(os.Getenv("BACKEND_URLS")), ",")
	healthCheckIntervalEnv := os.Getenv("HEALTH_CHECK_INTERVAL")

	var err error
	if healthCheckIntervalEnv != "" {
		_, err = strconv.Atoi(healthCheckIntervalEnv)
		if err != nil {
			log.Fatal("HEALTH_CHECK_INTERVAL enviornment variable is not valid")
		}
	}

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
	startTime := time.Now()

	// Create a logger with request-specific fields
	logger := log.WithFields(log.Fields{
		"method":     r.Method,
		"url":        r.URL.String(),
		"remoteAddr": r.RemoteAddr,
		"userAgent":  r.UserAgent(),
	})
	log.Printf("Incoming request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	loadBalance(w, r)

	// Create a custom ResponseWriter to capture response details
	crw := &customResponseWriter{ResponseWriter: w}

	logger.WithFields(log.Fields{
		"status":       crw.status,
		"responseTime": time.Since(startTime),
		"attempts":     attempts,
	}).Info("Request processed and response sent")
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
}

func (crw *customResponseWriter) WriteHeader(status int) {
	crw.status = status
	crw.ResponseWriter.WriteHeader(status)
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	if crw.status == 0 {
		crw.status = http.StatusOK
	}
	return crw.ResponseWriter.Write(b)
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
